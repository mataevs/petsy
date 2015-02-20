package role

type Sitter struct {
	baseRole
	Experience   string
	HousingType  string
	Space        string
	Price        int
	OwnsPets     bool
	HasCar       bool
	Permanent    bool
	ResponseRate float32
	ResponseTime float32
}
