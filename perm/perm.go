package perm

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	findPermStrings = regexp.MustCompile("(?U)(\\[(\\d|,|-)*\\])")
)

type permToken struct {
	Token string
	Perm  string
}

func getPermTokens(input string) ([]permToken, string) {
	tokens := []permToken{}
	ret := input
	next := findPermStrings.Find([]byte(input))
	cnt := 0
	for next != nil {
		cnt++
		token := "[%" + strconv.Itoa(cnt) + "]"
		perm := string(next[1 : len(next)-1])
		tokens = append(tokens, permToken{
			Token: token,
			Perm:  perm,
		})
		ret = strings.Replace(ret, string(next), token, 1)
		next = findPermStrings.Find([]byte(ret))
	}
	return tokens, ret
}

func parsePerm(perm string) []string {
	perm = strings.ReplaceAll(perm, " ", "")
	ret := []string{}
	commaSeparated := strings.Split(perm, ",")
	for _, val := range commaSeparated {
		minusSeparated := strings.Split(val, "-")
		if len(minusSeparated) == 1 {
			ret = append(ret, minusSeparated[0])
		} else if len(minusSeparated) == 2 {
			var begin, end int
			var err error
			if begin, err = strconv.Atoi(minusSeparated[0]); err != nil {
				ret = append(ret, val)
				continue
			}
			if end, err = strconv.Atoi(minusSeparated[1]); err != nil {
				ret = append(ret, val)
				continue
			}
			for i := begin; i <= end; i++ {
				ret = append(ret, strconv.Itoa(i))
			}
		} else {
			ret = append(ret, val)
		}
	}
	return ret
}

func Perm(input string) []string {
	tokens, input := getPermTokens(input)
	return permRecursive([]string{input}, tokens)
}

func permRecursive(inputs []string, perms []permToken) []string {
	if len(perms) == 0 {
		return inputs
	}
	cur := perms[0]
	permStrs := parsePerm(cur.Perm)
	next := []string{}
	for _, input := range inputs {
		for _, perm := range permStrs {
			next = append(next, strings.Replace(input, cur.Token, perm, 1))
		}
	}
	return permRecursive(next, perms[1:])
}
