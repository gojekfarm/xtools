package xkafka

import "time"

// ConsumerConfig are configurations for creating a Kafka consumer.
type ConsumerConfig struct {
	Topics             []string      `yaml:"TOPICS" env:"TOPICS"`
	Brokers            string        `yaml:"BROKERS" env:"BROKERS"`
	Group              string        `yaml:"GROUP" env:"GROUP"`
	TopicOffset        string        `yaml:"TOPIC_OFFSET" env:"TOPIC_OFFSET"`
	AutoCommit         bool          `yaml:"AUTO_COMMIT" env:"AUTO_COMMIT"`
	SessionTimeout     time.Duration `yaml:"SESSION_TIMEOUT" env:"SESSION_TIMEOUT"`
	HeartbeatInterval  time.Duration `yaml:"HEARTBEAT_INTERVAL" env:"HEARTBEAT_INTERVAL"`
	AutoCommitInterval time.Duration `yaml:"AUTO_COMMIT_INTERVAL" env:"AUTO_COMMIT_INTERVAL"`
	SocketTimeout      time.Duration `yaml:"SOCKET_TIMEOUT" env:"SOCKET_TIMEOUT"`
	ShutdownTimeout    time.Duration `yaml:"SHUTDOWN_TIMEOUT" env:"SHUTDOWN_TIMEOUT"`
	PollTimeout        time.Duration `yaml:"POLL_TIMEOUT" env:"POLL_TIMEOUT"`
	Concurrency        int           `yaml:"CONCURRENCY" env:"CONCURRENCY"`

	// Used with BatchConsumer
	BatchMaxSize int           `yaml:"MAX_SIZE" env:"MAX_SIZE"`
	BatchMaxWait time.Duration `yaml:"MAX_WAIT" env:"MAX_WAIT"`

	// Used as timeout while fetching metadata
	MetadataRequestTimeout int `yaml:"METADATA_REQUEST_TIMEOUT" env:"METADATA_REQUEST_TIMEOUT"`
}

// SetDefaults sets default values for ConsumerConfig.
func (c *ConsumerConfig) SetDefaults() {
	if c.TopicOffset == "" {
		c.TopicOffset = "latest"
	}

	if c.SessionTimeout == 0 {
		c.SessionTimeout = 10 * time.Second
	}

	if c.HeartbeatInterval == 0 {
		c.HeartbeatInterval = 3 * time.Second
	}

	if c.AutoCommitInterval == 0 {
		c.AutoCommitInterval = 5 * time.Second
	}

	if c.SocketTimeout == 0 {
		c.SocketTimeout = 10 * time.Second
	}

	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout = 10 * time.Second
	}

	if c.PollTimeout == 0 {
		c.PollTimeout = 10 * time.Second
	}

	if c.Concurrency == 0 {
		c.Concurrency = 1
	}

	if c.BatchMaxSize == 0 {
		c.BatchMaxSize = 100
	}

	if c.BatchMaxWait == 0 {
		c.BatchMaxWait = 100 * time.Millisecond
	}

	if c.MetadataRequestTimeout == 0 {
		c.MetadataRequestTimeout = 10000
	}
}

// ProducerConfig are configurations for creating a Kafka Producer.
type ProducerConfig struct {
	Brokers              string        `yaml:"BROKERS" env:"BROKERS"`
	ClientID             string        `yaml:"CLIENT_ID" env:"CLIENT_ID"`
	BatchNumMessages     int           `yaml:"BATCH_NUM_MESSAGES" env:"BATCH_NUM_MESSAGES"`
	Linger               time.Duration `yaml:"LINGER" env:"LINGER"`
	MaxRetries           int           `yaml:"MAX_RETRIES" env:"MAX_RETRIES"`
	RetryBackoff         time.Duration `yaml:"RETRY_BACKOFF" env:"RETRY_BACKOFF"`
	MessageTimeout       time.Duration `yaml:"MESSAGE_TIMEOUT" env:"MESSAGE_TIMEOUT"`
	StatsCollectInterval time.Duration `yaml:"STATS_COLLECT_INTERVAL" env:"STATS_COLLECT_INTERVAL"`
}
