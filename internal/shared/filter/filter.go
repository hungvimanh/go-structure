// Package filter provides reusable typed filter operators for entity search APIs.
package filter

type StringFilter struct {
	Eq         *string  `json:"eq"`
	Neq        *string  `json:"neq"`
	Contains   *string  `json:"contains"`
	StartsWith *string  `json:"startsWith"`
	EndsWith   *string  `json:"endsWith"`
	In         []string `json:"in"`
	NotIn      []string `json:"notIn"`
}

type IntFilter struct {
	Eq    *int  `json:"eq"`
	Neq   *int  `json:"neq"`
	Gt    *int  `json:"gt"`
	Gte   *int  `json:"gte"`
	Lt    *int  `json:"lt"`
	Lte   *int  `json:"lte"`
	In    []int `json:"in"`
	NotIn []int `json:"notIn"`
}
