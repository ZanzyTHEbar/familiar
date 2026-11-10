/* Linux scoped-search launcher. Uses Linux UAPI definitions (GPL-2.0 WITH
 * Linux-syscall-note); implementation is original Familiar code.
 * This executable is statically linked and single-threaded. No Go/FFF is loaded
 * until after restrict_self. FDs 3/4 are the sole granted directory capabilities.
 */
#define _GNU_SOURCE
#include <elf.h>
#include <errno.h>
#include <fcntl.h>
#include <limits.h>
#include <linux/landlock.h>
#include <linux/openat2.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/prctl.h>
#include <sys/stat.h>
#include <sys/syscall.h>
#include <unistd.h>

#if !defined(__x86_64__)
#error "This scoped-worker distribution currently supports Linux amd64 only"
#endif

static void die(const char *what) { perror(what); exit(125); }
static void rule(int ruleset,int fd,__u64 access) {
    struct landlock_path_beneath_attr attr={.allowed_access=access,.parent_fd=fd};
    if(syscall(SYS_landlock_add_rule,ruleset,LANDLOCK_RULE_PATH_BENEATH,&attr,0)) die("landlock_add_rule");
}
static void same_directory(int fd,const char *path) {
    struct stat given,named;
    if(fstat(fd,&given)||stat(path,&named)) die("grant identity");
    if(!S_ISDIR(given.st_mode)||given.st_dev!=named.st_dev||given.st_ino!=named.st_ino) { errno=ESTALE; die("grant changed"); }
}
static int runtime_file(int bundle,const char *name,int flags) {
    struct open_how how={.flags=flags|O_CLOEXEC,.resolve=RESOLVE_BENEATH|RESOLVE_NO_SYMLINKS};
    int fd=syscall(SYS_openat2,bundle,name,&how,sizeof how);
    if(fd<0) die("open runtime file beneath bundle");
    struct stat st;
    if(fstat(fd,&st)||!S_ISREG(st.st_mode)) { errno=EINVAL; die("runtime must be regular file"); }
    return fd;
}
static int interpreter(int bundle) {
    int worker=runtime_file(bundle,"worker",O_RDONLY);
    Elf64_Ehdr elf;
    if(pread(worker,&elf,sizeof elf,0)!=(ssize_t)sizeof elf || memcmp(elf.e_ident,ELFMAG,SELFMAG) ||
       elf.e_ident[EI_CLASS]!=ELFCLASS64 || elf.e_machine!=EM_X86_64 ||
       elf.e_phentsize!=sizeof(Elf64_Phdr) || elf.e_phnum>128) { errno=EINVAL; die("worker ELF header"); }
    for(unsigned int i=0;i<elf.e_phnum;i++) {
        Elf64_Phdr ph;
        if(pread(worker,&ph,sizeof ph,elf.e_phoff+i*sizeof ph)!=(ssize_t)sizeof ph) die("worker ELF program header");
        if(ph.p_type==PT_INTERP) {
            char path[PATH_MAX];
            if(ph.p_filesz<2 || ph.p_filesz>sizeof path ||
               pread(worker,path,ph.p_filesz,ph.p_offset)!=(ssize_t)ph.p_filesz ||
               path[0]!='/' || path[ph.p_filesz-1]!=0) { errno=EINVAL; die("worker ELF interpreter"); }
            close(worker);
            int fd=open(path,O_PATH|O_CLOEXEC);
            struct stat st;
            if(fd<0||fstat(fd,&st)||!S_ISREG(st.st_mode)) die("worker interpreter");
            return fd;
        }
    }
    errno=EINVAL; die("worker has no ELF interpreter"); return -1;
}
int main(int argc,char **argv) {
    if(argc!=5) { errno=EINVAL; die("launcher arguments"); }
    /* LD_LIBRARY_PATH interprets list separators and dollar tokens. Check here
     * as well as in the public API, before any worker/policy setup. */
    if(strpbrk(argv[1],":;$")) { errno=EINVAL; die("runtime path contains loader separator or token"); }
    for(int fd=0;fd<3;fd++) {
        struct stat st;
        if(fstat(fd,&st)||!S_ISFIFO(st.st_mode)) { errno=EINVAL; die("IPC must use pipes"); }
    }
    same_directory(3,argv[2]); same_directory(4,argv[3]);
    int abi=syscall(SYS_landlock_create_ruleset,NULL,0,LANDLOCK_CREATE_RULESET_VERSION);
    if(abi<5) { errno=ENOTSUP; die("Landlock ABI >= 5 (TRUNCATE and IOCTL_DEV) required"); }
    __u64 handled=(1ULL<<15)-1; /* all filesystem rights through ABI 3 */
    handled|=LANDLOCK_ACCESS_FS_IOCTL_DEV;
    struct landlock_ruleset_attr attr={.handled_access_fs=handled};
    int ruleset=syscall(SYS_landlock_create_ruleset,&attr,sizeof(__u64),0);
    if(ruleset<0) die("landlock_create_ruleset");
    rule(ruleset,3,LANDLOCK_ACCESS_FS_READ_FILE|LANDLOCK_ACCESS_FS_READ_DIR);
    /* State: ordinary files/dirs only. No devices, sockets, FIFOs or executables. */
    __u64 state_access=LANDLOCK_ACCESS_FS_READ_FILE|LANDLOCK_ACCESS_FS_READ_DIR|
        LANDLOCK_ACCESS_FS_WRITE_FILE|LANDLOCK_ACCESS_FS_TRUNCATE|
        LANDLOCK_ACCESS_FS_REMOVE_FILE|LANDLOCK_ACCESS_FS_REMOVE_DIR|
        LANDLOCK_ACCESS_FS_MAKE_REG|LANDLOCK_ACCESS_FS_MAKE_DIR|LANDLOCK_ACCESS_FS_REFER;
    rule(ruleset,4,state_access);
    char worker[PATH_MAX],line[PATH_MAX];
    if(snprintf(worker,sizeof worker,"%s/worker",argv[1])>=(int)sizeof worker) { errno=ENAMETOOLONG; die("runtime path"); }
    int bundle=open(argv[1],O_PATH|O_DIRECTORY|O_CLOEXEC|O_NOFOLLOW);
    if(bundle<0) die("runtime bundle");
    FILE *list=fdopen(runtime_file(bundle,"runtime.files",O_RDONLY),"r");
    if(!list) die("runtime manifest");
    int count=0;
    while(fgets(line,sizeof line,list)) {
        size_t len=strlen(line);
        if(!len||line[len-1]!='\n') { errno=EINVAL; die("runtime manifest line"); }
        line[--len]=0;
        if(!len) { errno=EINVAL; die("empty runtime entry"); }
        int is_interpreter=!strcmp(line,"@interpreter");
        int is_worker=!strcmp(line,"worker");
        if(!is_interpreter && !is_worker && (strncmp(line,"lib/",4)||strstr(line,"..")||strchr(line+4,'/'))) {
            errno=EINVAL; die("only bundled libraries and ELF interpreter are runtime resources");
        }
        int fd=is_interpreter ? interpreter(bundle) : runtime_file(bundle,line,O_PATH);
        rule(ruleset,fd,LANDLOCK_ACCESS_FS_READ_FILE|((is_interpreter||is_worker)?LANDLOCK_ACCESS_FS_EXECUTE:0));
        close(fd); count++;
    }
    if(ferror(list)||!count) { errno=EINVAL; die("runtime manifest read"); }
    fclose(list);
    if(chdir(argv[3])) die("state cwd");
    if(prctl(PR_SET_NO_NEW_PRIVS,1,0,0,0)) die("no_new_privs");
    if(syscall(SYS_landlock_restrict_self,ruleset,0)) die("landlock_restrict_self");
    /* Includes root/state/ruleset and any unexpected caller capabilities. */
    if(syscall(SYS_close_range,3,~0U,0)) die("close_range");
    char ldpath[PATH_MAX+32];
    if(snprintf(ldpath,sizeof ldpath,"LD_LIBRARY_PATH=%s/lib",argv[1])>=(int)sizeof ldpath) die("library path");
    char tmp[PATH_MAX+16];
    if(snprintf(tmp,sizeof tmp,"TMPDIR=%s",argv[3])>=(int)sizeof tmp) die("temp path");
    char *env[]={"FAMILIAR_SCOPED_CHILD=1","LANG=C","HOME=/nonexistent","XDG_CONFIG_HOME=/nonexistent",
        "PATH=/nonexistent","GOMAXPROCS=2","GOMEMLIMIT=256MiB","OMP_NUM_THREADS=2",
        "OPENBLAS_NUM_THREADS=2","RAYON_NUM_THREADS=2",ldpath,tmp,NULL};
    char *args[]={worker,argv[2],argv[3],argv[4],NULL};
    execve(worker,args,env);
    die("exec worker");
}
