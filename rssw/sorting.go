package rssw

//Allow sorting
type SortableItems []ItemObject

func (s SortableItems) Len() int      { return len(s) }
func (s SortableItems) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByTime struct{ SortableItems }

func (s ByTime) Less(i, j int) bool {
	return s.SortableItems[i].UnixTime() > s.SortableItems[j].UnixTime()
}
