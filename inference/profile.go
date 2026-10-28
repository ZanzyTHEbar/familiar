// Package inference supplies opt-in local inference boundaries. Nothing in this
// package is wired into doctor or starts a process during initialization.
package inference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// Identity is published candidate metadata, not proof of a locally verified or
// qualified artifact. LFM3+ means a reviewed pin, never a moving version.
type Identity struct {
	Role, Repository, Revision, File, SHA256 string
	Bytes                                    int64
	RuntimeRevision                          string
	Tokenizer, Template, ToolParser          string
	ContextTokens, Dimensions                int
	Pooling, Projection, Normalization       string
	QueryPrefix, DocumentPrefix              string
	Conversion, Chunking                     string
}

// Profile has no mutable exported state. Identity returns a value copy.
type Profile struct{ identity Identity }

func (p Profile) Identity() Identity { return p.identity }

func (p Profile) Fingerprint() string {
	b, _ := json.Marshal(p.identity)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

const runtimeRevision = "8642541b255bfb5dc39cd351a178be09013a55d6"

// GenerationCandidate is deliberately not a production-enabled default.
func GenerationCandidate() Profile {
	return Profile{Identity{
		Role: "generate", Repository: "LiquidAI/LFM3-1.7B-GGUF",
		Revision: "b3f2ae5d73cd3c91bf40e8ba752c648aa05dffdf", File: "LFM3-1.7B-Q4_K_M.gguf",
		SHA256: "3793e1e6e2266919d6e015287221b41100175a98eddaba92e111aed3b05c93c4", Bytes: 1099627808,
		RuntimeRevision: runtimeRevision, ContextTokens: 8192,
		Tokenizer: "embedded-in-artifact", Template: "embedded-in-artifact;enable_thinking=false",
		ToolParser: "runtime-normalized-function-calls;one-call;flat-string-object-v1",
	}}
}

// EmbeddingCandidate prefers a trained Liquid-family dense bi-encoder, not
// hidden-state pooling of a generator. This is metadata, not an Embed adapter.
func EmbeddingCandidate() Profile {
	return Profile{Identity{
		Role: "embed", Repository: "LiquidAI/LFM2.5-Embedding-350M-GGUF",
		Revision: "a80de9c5b941d429104f0038292a0ef5a860e486", File: "LFM2.5-Embedding-350M-Q8_0.gguf",
		SHA256: "6ec5f8e8750dbc8a0e40c431fd1b7b07a13688136b2244c5a1364b54d9032599", Bytes: 379216640,
		RuntimeRevision: runtimeRevision, ContextTokens: 512, Dimensions: 1024,
		Tokenizer: "embedded-in-artifact;BOS=1;no-EOS;pad=0;case-preserved", Template: "raw-prefixes;no-chat-template;cosine-v1",
		Pooling: "CLS", Projection: "none", Normalization: "L2",
		QueryPrefix: "query: ", DocumentPrefix: "document: ",
		Conversion: "source-revision-unattested;reference=f35ae2c91d687658dbf1f2b449382f0b019b9808",
		Chunking:   "512-including-prefix-specials;no-truncation;caller-versioned-source-ranges",
	}}
}

// LateInteractionCandidate is conditional on judged ranking failures and separate
// conformance. It cannot implement the single-vector Embedder port or be used as
// a stock cross-encoder. The stricter documented 511-token cap is deliberate.
func LateInteractionCandidate() Profile {
	return Profile{Identity{
		Role: "late-interaction-retrieval", Repository: "LiquidAI/LFM2.5-ColBERT-350M-GGUF",
		Revision: "bc240003aba07253e261a8aaf0d2c9683318a967", File: "LFM2.5-ColBERT-350M-Q8_0.gguf",
		SHA256: "5765e4987f1e63e6c52572198abd65fe9616416edf523c0fa95a077a3e501d47", Bytes: 378097248,
		RuntimeRevision: runtimeRevision, ContextTokens: 511, Dimensions: 128,
		Tokenizer: "embedded-in-artifact;reference-pad-and-expansion-mask-conformance-required",
		Template:  "raw-prefixes;query<=32;pad-to-32;reference-skiplist;MaxSim=sum-query-max-document-dot",
		Pooling:   "none;per-token", Projection: "learned-1024-to-128;no-bias;identity", Normalization: "per-token-L2",
		QueryPrefix: "[Q] ", DocumentPrefix: "[D] ",
		Conversion: "source-revision-unattested;reference=9772bdf797255d8693b83e84aa98e9b2d36dd0be",
		Chunking:   "511-including-prefix-specials;no-truncation;caller-versioned-source-ranges",
	}}
}

// EmbeddingFallbackCandidate is considered only if no suitable Liquid integration
// can be qualified. No automatic fallback or adapter is implemented.
func EmbeddingFallbackCandidate() Profile {
	return Profile{Identity{
		Role: "embed", Repository: "ggml-org/embeddinggemma-2-GGUF",
		Revision: "58faf90940886df2d62be555c8765d00dac362c5", File: "embeddinggemma-2-Q8_0.gguf",
		SHA256: "2188ac1deca4b77dffefd603c2776a9d76d9d74ec01841392982ebb840b09135", Bytes: 309855456,
		RuntimeRevision: runtimeRevision, ContextTokens: 8192, Dimensions: 768,
		Tokenizer: "embedded-in-artifact", Template: "explicit-text-prefixes-v1",
		Pooling: "mean", Projection: "required-learned-512-to-768", Normalization: "L2",
		QueryPrefix: "task: search result | query: ", DocumentPrefix: "title: {title-or-none} | text: ",
		Conversion: "914f7f89142e33e77833254d9c9b90c3cef7303b", Chunking: "no-truncation;caller-versioned-source-ranges",
	}}
}
