package suite

import (
	"fmt"
)

func scopeGlouton(name string, input []byte, d Defaults) (string, map[string]any, error) {
	switch name {
	case "glouton_create_company":
		return scopeDecode(name, input, d, func(in gloutonCreateCompanyInput, _ Defaults) (string, map[string]any) {
			return "companies", map[string]any{"resource": "company", "action": "create", "name": in.Name}
		})
	case "glouton_update_company":
		return scopeDecode(name, input, d, func(in gloutonUpdateCompanyInput, _ Defaults) (string, map[string]any) {
			return scopeValue("company", in.CompanyID), map[string]any{"resource": "company", "action": "update", "company": in.CompanyID}
		})
	case "glouton_create_person":
		return scopeDecode(name, input, d, func(in GloutonPersonFields, _ Defaults) (string, map[string]any) {
			company := ""
			if in.CompanyID != nil {
				company = *in.CompanyID
			}
			scope := "people"
			if company != "" {
				scope = "company:" + company
			}
			return scope, map[string]any{"resource": "person", "action": "create", "company": company}
		})
	case "glouton_update_person":
		return scopeDecode(name, input, d, func(in gloutonUpdatePersonInput, _ Defaults) (string, map[string]any) {
			return scopeValue("person", in.PersonID), map[string]any{"resource": "person", "action": "update", "person": in.PersonID}
		})
	}
	return "", nil, fmt.Errorf("suite: no scope rule for %q", name)
}
