// Package scopedwire is the private, single-client scoped-search pipe protocol.
package scopedwire

import "github.com/ZanzyTHEbar/virtual-vectorfs/search"

type Request struct {
	Op        string
	Documents []search.Document
	Query     search.Query
	Expand    search.ExpandRequest
}

type Response struct {
	Error   string
	Hits    []search.Hit
	Excerpt search.Excerpt
}
