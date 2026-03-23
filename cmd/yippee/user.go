package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
	"github.com/h1divp/yippee/internal/config"
	"github.com/h1divp/yippee/internal/store"
	"github.com/urfave/cli/v3"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#A78BFA")).
			PaddingLeft(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#34D399")).
			PaddingLeft(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			Width(12).
			Align(lipgloss.Right)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	adminBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FCD34D")).
			Background(lipgloss.Color("#78350F")).
			Padding(0, 1)

	userBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#93C5FD")).
			Background(lipgloss.Color("#1E3A5F")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6D28D9")).
			Padding(1, 3).
			MarginTop(1).
			MarginBottom(1)
)

func UserCommands() *cli.Command {
	return &cli.Command{

		Name:        "user",
		Usage:       "user commands",
		Description: "commands used when interfacing with users",
		Commands: []*cli.Command{
			{
				Name:        "create",
				Description: "create a new user",
				Usage:       "creates a user and scaffolds out the file structure",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "username",
						Usage:    "username for the new user",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "password",
						Usage:    "password for new user, prefer interactive prompt over password flag",
						Required: false,
					},
					&cli.BoolFlag{
						Name:     "admin",
						Usage:    "set this flag to create admin user",
						Value:    false,
						Required: false,
						Aliases:  []string{"a"},
					},
				},
				Action: createUser,
			},
		},
	}
}

func createUser(ctx context.Context, cmd *cli.Command) error {
	// First ensure filesystem is scaffolded
	basePath, ok := config.ValidateStructure()
	if !ok {
		fmt.Println("Yippee not initialized. Please run 'yippee init' before creating users.")
		return nil
	}
	// Then ensure DB can connect
	_, err := store.New(basePath)
	if err != nil {
		fmt.Errorf("something went wrong connecting to sqlite: %w", err)
		return err
	}

	username, password, isAdmin := cmd.String("username"), cmd.String("password"), cmd.Bool("admin")

	var fields []huh.Field

	if strings.TrimSpace(username) == "" {
		fields = append(fields,
			huh.NewInput().
				Inline(true).
				Title("Username ").
				Placeholder("e.g. johndoe").
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return errors.New("username cannot be empty")
					}
					if len(strings.TrimSpace(s)) < 3 {
						return errors.New("at least 3 characters")
					}
					return nil
				}).
				Value(&username),
		)
	}

	if strings.TrimSpace(password) == "" {
		fields = append(fields,
			huh.NewInput().
				Inline(true).
				Title("Password ").
				Placeholder("min 8 characters").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if len(s) < 8 {
						return errors.New("at least 8 characters")
					}
					return nil
				}).
				Value(&password),
		)
	}

	if !isAdmin {
		fields = append(fields,
			huh.NewConfirm().
				Inline(true).
				Title("Make user admin? ").
				Affirmative("Yes").
				Negative("No").
				Value(&isAdmin),
		)
	}

	if len(fields) > 0 {
		lipgloss.Println(titleStyle.Render("\n✦ New User"))

		form := huh.NewForm(
			huh.NewGroup(fields...),
		)

		err := form.Run()
		if err != nil {
			return err
		}
	}

	// ─── TODO: hash password, sqlite insert, mkdir ───

	// Role badge
	roleBadge := userBadge.Render("user")
	if isAdmin {
		roleBadge = adminBadge.Render("admin")
	}

	summary := strings.Join([]string{
		labelStyle.Render("Username") + "  " + valueStyle.Render(username),
		labelStyle.Render("Role") + "  " + roleBadge,
		labelStyle.Render("Home") + "  " + valueStyle.Render("~/.yippee/data/"+username),
	}, "\n")

	lipgloss.Println(boxStyle.Render(
		successStyle.Render("✔ User created") + "\n\n" + summary,
	))

	return nil
}
