package ellipxobj

type PairName [2]string

func Pair(a, b string) PairName {
	return PairName{a, b}
}
