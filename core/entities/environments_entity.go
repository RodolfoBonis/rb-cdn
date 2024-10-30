package entities

type environmentsEntity struct {
	Development string
	Staging     string
	Production  string
}

var Environment = environmentsEntity{
	Development: "development",
	Staging:     "staging",
	Production:  "production",
}
