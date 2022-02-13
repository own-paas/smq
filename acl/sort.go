package acl

import (
	"github.com/sestack/smq/model"
	"sort"
)

type byPriority []*model.Acl

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func sortByPriority(classify string, acls map[string]*model.Acl) []*model.Acl {
	aclList := []*model.Acl{}

	for _, acl := range acls {
		if acl.Classify == classify {
			aclList = append(aclList, acl)
		}
	}

	sort.Sort(byPriority(aclList))

	return aclList
}
