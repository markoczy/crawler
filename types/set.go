package types

type StringSet map[string]bool

func NewStringSet() *StringSet {
	set := StringSet(make(map[string]bool))
	return &set
}

func (set *StringSet) Add(s ...string) {
	for _, cur := range s {
		(*set)[cur] = true
	}
}

func (set *StringSet) Remove(s string) {
	delete(*set, s)
}

func (set *StringSet) Exists(s string) bool {
	return (*set)[s]
}

func (set *StringSet) Values() []string {
	ret := []string{}
	for k := range *set {
		ret = append(ret, k)
	}
	return ret
}

func (set *StringSet) Copy() *StringSet {
	ret := NewStringSet()
	for k := range *set {
		ret.Add(k)
	}
	return ret
}

func (set *StringSet) Len() int {
	return len(*set)
}
