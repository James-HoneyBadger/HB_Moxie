package _default

// defaultTransactionDataTSVFileConverter defines the structure of Honey Ledger default tsv file converter
type defaultTransactionDataTSVFileConverter struct {
	defaultTransactionDataPlainTextConverter
}

// Initialize a Honey Ledger default transaction data tsv file converter singleton instance
var (
	DefaultTransactionDataTSVFileConverter = &defaultTransactionDataTSVFileConverter{
		defaultTransactionDataPlainTextConverter{
			columnSeparator: "\t",
		},
	}
)
