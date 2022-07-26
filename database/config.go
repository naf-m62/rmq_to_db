package database

type Config struct {
	Host       string `yaml:"host"`
	Port       string `yaml:"port"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	DBName     string `yaml:"dbname"`
	TableName  string `yaml:"tableName"`
	DriverName string `yaml:"driverName"`
}
