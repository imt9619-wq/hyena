package pathfind

type Set []*node

func (s Set) Len() int{
	return len(s) 
}

func (s Set) Less(i, j int) bool{
	return s[i].f < s[j].f
} 

func (s Set) Swap(i, j int){
	s[i], s[j] = s[j], s[i] 
}

func (s *Set) Push(x any) {
	*s = append(*s, x.(*node))
}

func (s *Set) Pop() any{
	old := *s
	n := len(old)
	x := old[n-1]
	old[n-1] = nil
	*s = old[0 : n-1]
	return x
}