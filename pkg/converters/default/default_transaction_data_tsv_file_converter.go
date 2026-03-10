package _default

// defaultTransactionDataTSVFileConverter defines the structure of hbmoxie default tsv file converter
type defaultTransactionDataTSVFileConverter struct {
	defaultTransactionDataPlainTextConverter
}

// Initialize an hbmoxie default transaction data tsv file converter singleton instance
var (
	DefaultTransactionDataTSVFileConverter = &defaultTransactionDataTSVFileConverter{
		defaultTransactionDataPlainTextConverter{
			columnSeparator: "\t",
		},
	}
)
