package document

// Document wraps a html document (page) so that we can extract all the relavent
// urls from it.
type Document struct {
}

// NewDocument creates a new Document to use.
func NewDocument() *Document {
	return &Document{}
}
