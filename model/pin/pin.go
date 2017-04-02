// Package pin has types for describing pins (connection points).
package pin

// Direction describes which way information flows in a Pin.
type Direction string

// The various directions.
const (
	Input  Direction = "in"
	Output Direction = "out"
)

// Definition describes the main properties of a pin
type Definition struct {
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Direction Direction `json:"dir"`
}

// FullType returns the full pin type, including the <-chan / chan<-.
func (d *Definition) FullType() string {
	c := "<-chan "
	if d.Direction == Output {
		c = "chan<- "
	}
	return c + d.Type
}
