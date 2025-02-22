package cmd

//"github.com/apex/log"

type Hub struct {
	Url      string `yaml:"url" mapstructure:"url" json:"url" validate:"required"`
	ApiKey   string `yaml:"api_key,omitempty" mapstructure:"api_key,omitempty" json:"api_key,omitempty"`
	Username string `yaml:"username,omitempty" mapstructure:"username,omitempty" json:"username,omitempty"`
	Password string `yaml:"password,omitempty" mapstructure:"password,omitempty" json:"password,omitempty"`
}
