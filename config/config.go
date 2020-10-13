package config

import (
  "errors"
  "gopkg.in/yaml.v2"
  "log"
  "os"
)

var (
  ErrNoConfig = errors.New("configuration not loaded")
  ErrAuth     = errors.New("username/password did not match to any user")

  cfg *Config
)

type Config struct {
  Users      []User `yaml:"users"`
  Nameserver string `yaml:"nameserver"`
  Mailbox    string `yaml:"mailbox"`
}

type User struct {
  Zones    []string `yaml:"zones"`
  Username string   `yaml:"username,omitempty"`
  Password string   `yaml:"password,omitempty"`
}

func LoadConfig(path string) error {

  file, err := os.Open(path)
  if err != nil {
    return err
  }

  c := &Config{}
  dec := yaml.NewDecoder(file)
  err = dec.Decode(c)
  if err != nil {
    return err
  }
  cfg = c
  return nil
}

func GetNameserver() string {
  if cfg == nil || cfg.Nameserver == "" {
    log.Fatal(errors.New("config missing nameserver"))
  }
  return cfg.Nameserver
}
func GetMailbox() string {
  if cfg == nil || cfg.Mailbox == "" {
    return "hostmaster." + GetNameserver()
  }
  return cfg.Mailbox
}

func GetUserBasic(username, password string) (*User, error) {
  if cfg == nil {
    return nil, ErrNoConfig
  }
  if username == "" || password == "" {
    return nil, ErrAuth
  }
  for _, u := range cfg.Users {
    if u.Username == username && u.Password == password {
      return &u, nil
    }
  }
  return nil, ErrAuth
}
