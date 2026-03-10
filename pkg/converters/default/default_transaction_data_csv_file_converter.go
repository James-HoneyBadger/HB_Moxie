package _default

// defaultTransactionDataCSVFileConverter defines the structure of hbmoxie default csv file converter
type defaultTransactionDataCSVFileConverter struct {
	defaultTransactionDataPlainTextConverter
}

// Initialize an hbmoxie default transaction data csv file converter singleton instance
var (
	DefaultTransactionDataCSVFileConverter = &defaultTransactionDataCSVFileConverter{
		defaultTransactionDataPlainTextConverter{
			columnSeparator: ",",
		},
	}
)
