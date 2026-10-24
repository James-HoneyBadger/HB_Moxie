package cmd

import (
	"encoding/json"
	"os"

	"github.com/James-HoneyBadger/ratel-ledger/pkg/avatars"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/core"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/datastore"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/duplicatechecker"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/exchangerates"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/llm"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/log"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/mail"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/models"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/services"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/settings"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/storage"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/utils"
	"github.com/James-HoneyBadger/ratel-ledger/pkg/uuid"
)

func initializeSystem(c *core.CliContext) (*settings.Config, error) {
	var err error
	configFilePath := c.String("conf-path")
	isDisableBootLog := c.Bool("no-boot-log")

	if configFilePath != "" {
		if _, err = os.Stat(configFilePath); err != nil {
			if !isDisableBootLog {
				log.BootErrorf(c, "[initializer.initializeSystem] cannot load configuration from custom config path %s, because file not exists", configFilePath)
			}
			return nil, err
		}

		if !isDisableBootLog {
			log.BootInfof(c, "[initializer.initializeSystem] will loading configuration from custom config path %s", configFilePath)
		}
	} else {
		configFilePath, err = settings.GetDefaultConfigFilePath()

		if err != nil {
			if !isDisableBootLog {
				log.BootErrorf(c, "[initializer.initializeSystem] cannot get default configuration path, because %s", err.Error())
			}
			return nil, err
		}

		if !isDisableBootLog {
			log.BootInfof(c, "[initializer.initializeSystem] will load configuration from default config path %s", configFilePath)
		}
	}

	config, err := settings.LoadConfiguration(configFilePath)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] cannot load configuration, because %s", err.Error())
		}
		return nil, err
	}

	if config.SecretKeyNoSet {
		// Auto-persist the generated secret key so tokens survive restarts
		if err := settings.PersistSecretKey(configFilePath, config.SecretKey); err != nil {
			if !isDisableBootLog {
				log.BootWarnf(c, "[initializer.initializeSystem] failed to persist auto-generated secret_key: %s", err.Error())
				log.BootWarnf(c, "[initializer.initializeSystem] \"secret_key\" in config file is not set, please change it to keep your user data safe")
			}
		} else {
			config.SecretKeyNoSet = false
			if !isDisableBootLog {
				log.BootInfof(c, "[initializer.initializeSystem] auto-generated and persisted secret_key to config file")
			}
		}
	}

	settings.SetCurrentConfig(config)

	err = datastore.InitializeDataStore(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes data store failed, because %s", err.Error())
		}
		return nil, err
	}

	err = log.SetLoggerConfiguration(config, isDisableBootLog)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] sets logger configuration failed, because %s", err.Error())
		}
		return nil, err
	}

	err = storage.InitializeStorageContainer(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes object storage failed, because %s", err.Error())
		}
		return nil, err
	}

	err = llm.InitializeLargeLanguageModelProvider(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes large language model provider failed, because %s", err.Error())
		}
		return nil, err
	}

	err = uuid.InitializeUuidGenerator(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes uuid generator failed, because %s", err.Error())
		}
		return nil, err
	}

	err = duplicatechecker.InitializeDuplicateChecker(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes duplicate checker failed, because %s", err.Error())
		}
		return nil, err
	}

	err = avatars.InitializeAvatarProvider(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes avatar provider failed, because %s", err.Error())
		}
		return nil, err
	}

	err = mail.InitializeMailer(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes mailer failed, because %s", err.Error())
		}
		return nil, err
	}

	err = exchangerates.InitializeExchangeRatesDataSource(config)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] initializes exchange rates data source failed, because %s", err.Error())
		}
		return nil, err
	}

	if !isDisableBootLog {
		if sanitizedConfig := getConfigWithoutSensitiveData(config); sanitizedConfig != nil {
			if cfgJson, marshalErr := json.Marshal(sanitizedConfig); marshalErr == nil {
				log.BootInfof(c, "[initializer.initializeSystem] has loaded configuration %s", cfgJson)
			}
		}
	}

	// The default-user lookup below queries the database, so the schema must
	// exist before single-user bootstrap runs on a fresh installation.
	if config.AutoUpdateDatabase {
		err = updateAllDatabaseTablesStructure(c)

		if err != nil {
			if !isDisableBootLog {
				log.BootErrorf(c, "[initializer.initializeSystem] update database table structure failed, because %s", err.Error())
			}
			return nil, err
		}
	}

	// Ensure a default user exists for single-user mode
	err = ensureDefaultUser(c, isDisableBootLog)

	if err != nil {
		if !isDisableBootLog {
			log.BootErrorf(c, "[initializer.initializeSystem] failed to ensure default user, because %s", err.Error())
		}
		return nil, err
	}

	return config, nil
}

func ensureDefaultUser(c *core.CliContext, isDisableBootLog bool) error {
	user, err := services.Users.GetFirstUser(c)

	if err != nil {
		return err
	}

	if user != nil {
		if !isDisableBootLog {
			log.BootInfof(c, "[initializer.ensureDefaultUser] default user \"%s\" (uid:%d) already exists", user.Username, user.Uid)
		}
		return nil
	}

	// No users exist — create the default single-user account
	generatedPassword, err := utils.GetRandomNumberOrLetter(16)

	if err != nil {
		return err
	}

	defaultUser := &models.User{
		Username:             "admin",
		Email:                "admin@ratel.local",
		Nickname:             "Ratel Ledger User",
		Password:             generatedPassword,
		DefaultCurrency:      "USD",
		FirstDayOfWeek:       core.WEEKDAY_SUNDAY,
		TransactionEditScope: models.TRANSACTION_EDIT_SCOPE_ALL,
		EmailVerified:        true,
	}

	err = services.Users.CreateUser(c, defaultUser, false)

	if err != nil {
		return err
	}

	if !isDisableBootLog {
		log.BootInfof(c, "[initializer.ensureDefaultUser] created default user \"%s\" (uid:%d)", defaultUser.Username, defaultUser.Uid)
		log.BootWarnf(c, "[initializer.ensureDefaultUser] IMPORTANT: default password is \"%s\" — please change it immediately", generatedPassword)
	}

	return nil
}

func getConfigWithoutSensitiveData(config *settings.Config) *settings.Config {
	if config == nil {
		return nil
	}

	clonedConfig := &settings.Config{}
	err := utils.Clone(config, clonedConfig)

	// Never fall back to the original configuration: it contains credentials.
	if err != nil {
		return nil
	}

	if clonedConfig.DatabaseConfig.DatabasePassword != "" {
		clonedConfig.DatabaseConfig.DatabasePassword = "****"
	}

	if clonedConfig.SMTPConfig.SMTPPasswd != "" {
		clonedConfig.SMTPConfig.SMTPPasswd = "****"
	}

	if clonedConfig.MinIOConfig.SecretAccessKey != "" {
		clonedConfig.MinIOConfig.SecretAccessKey = "****"
	}

	if clonedConfig.SecretKey != "" {
		clonedConfig.SecretKey = "****"
	}

	if clonedConfig.AmapApplicationSecret != "" {
		clonedConfig.AmapApplicationSecret = "****"
	}

	if clonedConfig.AmapApplicationKey != "" {
		clonedConfig.AmapApplicationKey = "****"
	}

	if clonedConfig.TomTomMapAPIKey != "" {
		clonedConfig.TomTomMapAPIKey = "****"
	}

	if clonedConfig.TianDiTuAPIKey != "" {
		clonedConfig.TianDiTuAPIKey = "****"
	}

	if clonedConfig.GoogleMapAPIKey != "" {
		clonedConfig.GoogleMapAPIKey = "****"
	}

	if clonedConfig.BaiduMapAK != "" {
		clonedConfig.BaiduMapAK = "****"
	}

	if clonedConfig.WebDAVConfig != nil && clonedConfig.WebDAVConfig.Password != "" {
		clonedConfig.WebDAVConfig.Password = "****"
	}

	if clonedConfig.ReceiptImageRecognitionLLMConfig != nil {
		if clonedConfig.ReceiptImageRecognitionLLMConfig.OpenAIAPIKey != "" {
			clonedConfig.ReceiptImageRecognitionLLMConfig.OpenAIAPIKey = "****"
		}

		if clonedConfig.ReceiptImageRecognitionLLMConfig.OpenAICompatibleAPIKey != "" {
			clonedConfig.ReceiptImageRecognitionLLMConfig.OpenAICompatibleAPIKey = "****"
		}

		if clonedConfig.ReceiptImageRecognitionLLMConfig.AnthropicCompatibleAPIKey != "" {
			clonedConfig.ReceiptImageRecognitionLLMConfig.AnthropicCompatibleAPIKey = "****"
		}

		if clonedConfig.ReceiptImageRecognitionLLMConfig.AnthropicAPIKey != "" {
			clonedConfig.ReceiptImageRecognitionLLMConfig.AnthropicAPIKey = "****"
		}

		if clonedConfig.ReceiptImageRecognitionLLMConfig.OpenRouterAPIKey != "" {
			clonedConfig.ReceiptImageRecognitionLLMConfig.OpenRouterAPIKey = "****"
		}

		if clonedConfig.ReceiptImageRecognitionLLMConfig.LMStudioToken != "" {
			clonedConfig.ReceiptImageRecognitionLLMConfig.LMStudioToken = "****"
		}

		if clonedConfig.ReceiptImageRecognitionLLMConfig.GoogleAIAPIKey != "" {
			clonedConfig.ReceiptImageRecognitionLLMConfig.GoogleAIAPIKey = "****"
		}
	}

	if clonedConfig.OAuth2ClientSecret != "" {
		clonedConfig.OAuth2ClientSecret = "****"
	}

	return clonedConfig
}
