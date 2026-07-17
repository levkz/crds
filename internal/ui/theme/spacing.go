package theme

type Spacing struct {
	Xxs int
	Xs  int
	Sm  int
	Md  int
	Lg  int
	Xl  int
	Xxl int
}

var DefaultSpacing = Spacing{
	Xxs: 2,
	Xs:  4,
	Sm:  8,
	Md:  16,
	Lg:  24,
	Xl:  32,
	Xxl: 48,
}
