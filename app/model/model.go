package model

type Attribute struct {
	Name            string `json:"name"`
	IsMultipleField string `json:"isMultipleField"`
	IsMandatory     string `json:"isMandatory"`
	DataType        string `json:"dataType"`
	Id              string `json:"id"`
}
