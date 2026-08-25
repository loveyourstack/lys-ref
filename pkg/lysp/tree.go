package lysp

// in 1.26 this is unusable because all JSON items are wrapped with Item's json tag.
// Try in go 1.27 with json/v2 - using the 'embed' tag should flatten the item.
type TreeNode[T any] struct {
	Item     T    `json:"embed"`
	Children []*T `json:"children,omitempty"`
}
