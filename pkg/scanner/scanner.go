package scanner

type Scanner struct {
}

type Config struct {
}

func New(config Config) (*Scanner, error) {

	s := Scanner{}
	return &s, nil
}

func (s *Scanner) Start() {

}
