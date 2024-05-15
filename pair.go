package ellipxobj

type PairName [2]string

func Pair(a, b string) PairName {
	return PairName{a, b}
}

func (p PairName) String() string {
	return p[0] + "_" + p[1]
}
