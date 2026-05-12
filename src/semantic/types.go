package semantic

type Type string

const (
	TypeInt     Type = "INT"
	TypeFloat   Type = "FLOAT"
	TypeString  Type = "STRING"
	TypeBoolean Type = "BOOLEAN"
	TypeNull    Type = "NULL"
	TypeUnknown Type = "UNKNOWN"
)

// String returns the string representation of the type.
func (t Type) String() string {
	return string(t)
}
