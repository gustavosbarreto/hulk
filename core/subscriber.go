package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/OSSystems/hulk/unit"
	interpol "github.com/imkira/go-interpol"
	"github.com/joho/godotenv"

	"io/ioutil"
)

type Subscriber struct {
	unit        *unit.Subscriber
	topics      []string
	extraTopics []string
	environment map[string]string
}

func NewSubscriber() *Subscriber {
	return &Subscriber{
		environment: make(map[string]string),
	}
}

func (s *Subscriber) LoadUnit(file string) error {
	data, err := ioutil.ReadFile(file)

	s.unit, err = unit.SubscriberFromData(data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Subscriber) Receiver(topic string, payload []byte) {
	s.ExecuteHook(s.unit.Hooks.OnPublish, payload)
}

func (s *Subscriber) LoadEnvironmentFiles() error {
	for _, file := range s.unit.EnvironmentFiles {
		err := s.LoadEnvironment(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Subscriber) LoadEnvironment(file string) error {
	env, err := godotenv.Read(file)
	if err != nil {
		return err
	}

	for key, value := range env {
		if _, ok := s.environment[key]; ok {
			return fmt.Errorf("Duplicated environment variable: %s", key)
		}

		s.environment[key] = value
	}

	return nil
}

func (s *Subscriber) ExpandTopics() error {
	s.topics = s.topics[:0]

	// Append unexpanded extraTopics to topics specified in unit file
	allTopics := s.unit.Topics
	allTopics = append(allTopics, s.extraTopics...)

	for _, topic := range allTopics {
		expanded, err := interpol.WithMap(topic, s.environment)
		if err != nil {
			return err
		}

		s.topics = append(s.topics, expanded)
	}

	return nil
}

func (s *Subscriber) GetTopics() []string {
	return s.topics
}

func (s *Subscriber) CreateHookCommand(cmdLine string) *exec.Cmd {
	args := strings.Split(cmdLine, " ")
	command := args[0]

	if len(args) > 1 {
		args = args[1:]
	}

	cmd := exec.Command(command, args...)

	for key, value := range s.environment {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	return cmd
}

func (s *Subscriber) ExecuteHook(cmdLine string, payload []byte) error {
	cmd := s.CreateHookCommand(cmdLine)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	defer stdin.Close()

	err = cmd.Start()
	if err != nil {
		return err
	}

	stdin.Write(payload)

	return nil
}

func LoadSubscribers(path string) ([]*Subscriber, error) {
	if stat, err := os.Stat(path); err != nil {
		return nil, err
	} else {
		if !stat.IsDir() {
			return nil, fmt.Errorf("Not a directory")
		}
	}

	files, err := filepath.Glob(filepath.Join(path, "*.yaml"))
	if err != nil {
		return nil, err
	}

	subscribers := []*Subscriber{}

	for _, file := range files {
		subscriber := NewSubscriber()

		if err := subscriber.LoadUnit(file); err != nil {
			return nil, err
		}

		subscribers = append(subscribers, subscriber)
	}

	return subscribers, nil
}
