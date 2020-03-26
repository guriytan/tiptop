package tiptop

func isPowerOfTwo(number int) bool {
	return (number & (number - 1)) == 0
}
