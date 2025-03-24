package broker

// Attributes is an interface for managing key-value pairs associated with a message.
type Attributes interface {
	// Add adds a key-value pair to the attributes.
	Add(key, value string)

	// Get gets the value of a key in the attributes.
	Get(key string) string

	// Lookup looks up the value of a key in the attributes.
	Lookup(key string) (string, bool)

	// Delete deletes a key-value pair from the attributes.
	Delete(key string)

	// Values returns a map of all key-value pairs in the attributes.
	Values() map[string][]string
}

// Message is an interface that represents a message, with a payload, attributes, and a unique identifier.
type Message interface {
	// Payload returns the payload of the message.
	Payload() any

	// Attributes returns the attributes associated with the message.
	Attributes() Attributes
}
