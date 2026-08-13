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

// Target sets or displays the currently targeted organization and space
// GUIDs. Mirrors `cf target -o ORG -s SPACE` so scripts can swap `cf`
// for `sb` and keep working. With no flags it just prints the current
// context.
func Target(cmd *Commandline) {
	c := Config{}
	c.load()

	changed := false
	if cmd.Organization != "" {
		c.OrganizationGUID = cmd.Organization
		changed = true
	}
	if cmd.Space != "" {
		c.SpaceGUID = cmd.Space
		changed = true
	}
	if changed {
		if err := c.save(); err != nil {
			CheckErr(err)
		}
	}

	if c.Host == "" {
		fmt.Println("No api endpoint set!")
		return
	}
	fmt.Printf("API endpoint:      %s\n", c.Host)
	fmt.Printf("User:              %s\n", c.Username)
	fmt.Printf("Organization GUID: %s\n", c.OrganizationGUID)
	fmt.Printf("Space GUID:        %s\n", c.SpaceGUID)
}

// CreateOrg is a no-op that logs a cf-shaped success line. The Service
// Broker has no concept of orgs; this exists so test scripts that call
// `cf create-org ORG` can be pointed at `sb` without editing.
func CreateOrg(cmd *Commandline) {
	if len(cmd.Options) < 1 {
		CheckErr(errors.New("Missing arguments!"), GetHelpText("create-org"))
	}
	who := currentUser()
	fmt.Printf("Creating org %s as %s...\n", cmd.Options[0], who)
	fmt.Println("OK")
}

// CreateSpace is a no-op that logs. Same rationale as CreateOrg.
func CreateSpace(cmd *Commandline) {
	if len(cmd.Options) < 1 {
		CheckErr(errors.New("Missing arguments!"), GetHelpText("create-space"))
	}
	who := currentUser()
	if cmd.Organization != "" {
		fmt.Printf("Creating space %s in org %s as %s...\n", cmd.Options[0], cmd.Organization, who)
	} else {
		fmt.Printf("Creating space %s as %s...\n", cmd.Options[0], who)
	}
	fmt.Println("OK")
}

func currentUser() string {
	conf := Config{}
	conf.load()
	if conf.Username == "" {
		return "-"
	}
	return conf.Username
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
		fmt.Printf("SkipSSLValidation: %t\n", conf.SkipSslValidation)
	}
}
