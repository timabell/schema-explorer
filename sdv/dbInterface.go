package sdv

type dbInterface interface{
	GetTables() (tables []TableName, err error)
}
