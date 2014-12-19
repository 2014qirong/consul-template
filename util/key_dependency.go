package util

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	api "github.com/armon/consul-api"
)

// from inside a template.
type KeyDependency struct {
	rawKey     string
	Path       string
	DataCenter string
}

// Fetch queries the Consul API defined by the given client and returns string
// of the value to Path.
func (d *KeyDependency) Fetch(client *api.Client, options *api.QueryOptions) (interface{}, *api.QueryMeta, error) {
	if d.DataCenter != "" {
		options.Datacenter = d.DataCenter
	}

	log.Printf("[DEBUG] (%s) querying consul with %+v", d.Display(), options)

	store := client.KV()
	pair, qm, err := store.Get(d.Path, options)
	if err != nil {
		return "", qm, err
	}

	if pair == nil {
		log.Printf("[DEBUG] (%s) Consul returned nothing (does the path exist?)",
			d.Display())
		return "", qm, nil
	}

	log.Printf("[DEBUG] (%s) Consul returned %s", d.Display(), pair.Value)

	return string(pair.Value), qm, nil
}

func (d *KeyDependency) HashCode() string {
	return fmt.Sprintf("KeyDependency|%s", d.Key())
}

func (d *KeyDependency) Key() string {
	return d.rawKey
}

func (d *KeyDependency) Display() string {
	return fmt.Sprintf(`key "%s"`, d.rawKey)
}

// AddToContext accepts a TemplateContext and data. It coerces the interface{}
// data into the correct format via type assertions, returning an errors that
// occur. The data is then set on the TemplateContext.
func (d *KeyDependency) AddToContext(context *TemplateContext, data interface{}) error {
	coerced, ok := data.(string)
	if !ok {
		return fmt.Errorf("key dependency: could not convert to string")
	}

	context.Keys[d.rawKey] = coerced
	return nil
}

// InContext checks if the dependency is contained in the given TemplateContext.
func (d *KeyDependency) InContext(c *TemplateContext) bool {
	_, ok := c.Keys[d.rawKey]
	return ok
}

func KeyFunc(deps map[string]Dependency) func(...string) (interface{}, error) {
	return func(s ...string) (interface{}, error) {
		if len(s) != 1 {
			return nil, fmt.Errorf("key: expected 1 argument, got %d", len(s))
		}

		d, err := ParseKeyDependency(s[0])
		if err != nil {
			return nil, err
		}

		if _, ok := deps[d.HashCode()]; !ok {
			deps[d.HashCode()] = d
		}

		return "", nil
	}
}

// ParseKeyDependency parses a string of the format a(/b(/c...))
func ParseKeyDependency(s string) (*KeyDependency, error) {
	if len(s) == 0 {
		return nil, errors.New("cannot specify empty key dependency")
	}

	re := regexp.MustCompile(`\A` +
		`(?P<key>[[:word:]\.\:\-\/]+)` +
		`(@(?P<datacenter>[[:word:]\.\-]+))?` +
		`\z`)
	names := re.SubexpNames()
	match := re.FindAllStringSubmatch(s, -1)

	if len(match) == 0 {
		return nil, errors.New("invalid key dependency format")
	}

	r := match[0]

	m := map[string]string{}
	for i, n := range r {
		if names[i] != "" {
			m[names[i]] = n
		}
	}

	key, datacenter := m["key"], m["datacenter"]

	if key == "" {
		return nil, errors.New("key part is required")
	}

	kd := &KeyDependency{
		rawKey:     s,
		Path:       key,
		DataCenter: datacenter,
	}

	return kd, nil
}
