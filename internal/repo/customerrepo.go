package repo

type Customer struct {
	Name           string
	Surname        string
	Phone          string
	MasterPassword string
}

type CustomerRepo struct {
	Customers []Customer
}

func CreateCustomer(name, surname, phone, masterPassword string) Customer {
	return Customer{Name: name, Surname: surname, Phone: phone, MasterPassword: masterPassword}
}

func (c *CustomerRepo) AddCustomer(cust Customer) {
	c.Customers = append(c.Customers, cust)
}
