package defs

// BasicAsset a basic asset
type BasicAsset struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Value     int    `json:"value"`
	Condition int    `json:"condition"`
}

// SetConditionNew set the condition of the asset to mark as new
func (ba *BasicAsset) SetConditionNew() {
	ba.Condition = 0
}

// SetConditionUsed set the condition of the asset to mark as used
func (ba *BasicAsset) SetConditionUsed() {
	ba.Condition = 1
}
