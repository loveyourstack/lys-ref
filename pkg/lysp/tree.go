package lysp

// TreeNode is a generic tree node item plus its children, nested to form a hierarchy
type TreeNode[T any] struct {
	Item     T              `json:"item"`
	Children []*TreeNode[T] `json:"children"`
}
