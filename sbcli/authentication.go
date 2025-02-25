package sbcli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

// retreieves all service instances from the service broker
func Logout(cmd *Commandline) {
	fmt.Println("Loggin out...")

	c := Config{}
	c.load()
	c.Password = ""
	c.Username = ""
	c.save()

	fmt.Println("OK")
}

func Api(cmd *Commandline) {
	c := Config{}
	c.load()

	if len(cmd.Options) == 0 {
		if c.Host == "" {
			fmt.Printf("No api endpoint set!\n")
		} else {
			fmt.Printf("API endpoint: %s\n", c.Host)
			fmt.Printf("User:         %s\n", c.Username)
		}
	} else {
		host := CleanTargetURI(cmd.Options[0])
		sb := NewSBClient(&Credentials{Host: host})
		err := sb.TestConnection()
		CheckErr(err)

		c.Host = host
		c.Password = ""
		c.Username = ""
		c.save()

		fmt.Printf("Target set to %s\n\n", c.Host)
		fmt.Printf("You have to login now.\n")
		fmt.Printf("\tsb login\n")
	}
}

func Auth(cmd *Commandline) {
	if len(cmd.Options) != 2 {
		CheckErr(errors.New("Missing arguments!"), GetHelpText("Auth"))
	}
	conf := Config{}
	conf.load()

	// check host
	if conf.Host == "" {
		CheckErr(errors.New("No target set."))
	}
	fmt.Printf("Target: %s...", conf.Host)

	// check if host is reachable
	sb := NewSBClient(&Credentials{Host: conf.Host})
	err := sb.TestConnection()
	CheckErr(err)

	fmt.Printf("OK\n\n")

	fmt.Printf("\nAuthenticating...")
	sb = NewSBClient(&Credentials{Host: conf.Host, Username: cmd.Options[0], Password: cmd.Options[1]})
	_, err = sb.Catalog()
	CheckErr(err)

	conf.Username = cmd.Options[0]
	conf.Password = cmd.Options[1]
	conf.save()

	fmt.Printf("OK\n\n")
}

func Login(cmd *Commandline) {
	conf := Config{}
	conf.load()

	if len(cmd.Api) > 0 {
		conf.Host = cmd.Api
	} else {
		// check host
		if conf.Host == "" {
			CheckErr(errors.New("No target set!"))
		}
		fmt.Printf("Target: %s...", conf.Host)
	}

	// check if host is reachable
	sb := NewSBClient(&Credentials{Host: conf.Host})
	err := sb.TestConnection()
	CheckErr(err)

	fmt.Printf("OK\n\n")

	c := Credentials{Host: conf.Host}

	if len(cmd.Username) > 0 {
		c.Username = cmd.Username
	} else {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Username> ")
		c.Username, _ = reader.ReadString('\n')
		c.Username = strings.TrimSpace(c.Username)

		if c.Username == "" {
			fmt.Printf("No username given, break!\n")
			os.Exit(1)
		}
	}

	fmt.Println()
	c.SkipSslValidation = cmd.SkipSslValidation

	ok := false
	for i := 0; i < 3; i++ {
		if i == 0 && len(cmd.Plan) > 0 {
			c.Password = cmd.Plan
		} else {
			c.Password, _ = getPassword("Password> ")
		}

		fmt.Printf("\nAuthenticating...")
		sb := NewSBClient(&c)
		_, err = sb.Catalog()
		if err != nil {
			fmt.Printf("Failed!\n\n")
			continue
		}
		fmt.Printf("OK\n\n")
		ok = true
		break
	}

	if ok {
		conf.Username = c.Username
		conf.Password = c.Password
		conf.SkipSslValidation = cmd.SkipSslValidation
		conf.save()
		fmt.Printf("Target:            %s\n", conf.Host)
		fmt.Printf("Username:          %s\n", conf.Username)
		fmt.Printf("SkipSSLValidation: %d\n", conf.SkipSslValidation)
	}
}
