package stnetbox

type Server struct {
	Name string
	IP   string
}

func GetInv() ([]Server, error) {

	return []Server{
		Server{"sls001", "10.1.1.1"},
		Server{"sls002", "10.1.1.2"},
		Server{"sls003", "10.1.1.3"},
		Server{"sls004", "10.1.1.3"},
		Server{"sls005", "10.1.1.3"},
		Server{"sls006", "10.1.1.3"},
		Server{"sls007", "10.1.1.3"},
		Server{"sls008", "10.1.1.3"},
		Server{"sls009", "10.1.1.3"},
		Server{"sls010", "10.1.1.3"},
		Server{"sls011", "10.1.1.1"},
		Server{"sls012", "10.1.1.2"},
		Server{"sls013", "10.1.1.3"},
		Server{"sls014", "10.1.1.3"},
		Server{"sls015", "10.1.1.3"},
		Server{"sls016", "10.1.1.3"},
		Server{"sls017", "10.1.1.3"},
		Server{"sls018", "10.1.1.3"},
		Server{"sls019", "10.1.1.3"},
		Server{"sls020", "10.1.1.3"},
		Server{"sls021", "10.1.1.1"},
		Server{"sls022", "10.1.1.2"},
		Server{"sls023", "10.1.1.3"},
		Server{"sls024", "10.1.1.3"},
		Server{"sls025", "10.1.1.3"},
		Server{"sls026", "10.1.1.3"},
		Server{"sls027", "10.1.1.3"},
		Server{"sls028", "10.1.1.3"},
		Server{"sls029", "10.1.1.3"},
		Server{"sls030", "10.1.1.3"},
	}, nil

}
