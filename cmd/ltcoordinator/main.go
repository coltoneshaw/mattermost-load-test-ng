// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mattermost/mattermost-load-test-ng/coordinator"
	"github.com/mattermost/mattermost-load-test-ng/loadtest"
	"github.com/mattermost/mattermost-load-test-ng/logger"

	"github.com/spf13/cobra"
)

func RunCoordinatorCmdF(cmd *cobra.Command, args []string) error {
	configFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	cfg, err := coordinator.ReadConfig(configFilePath)
	if err != nil {
		return err
	}

	log := logger.New(&cfg.LogSettings)

	ltConfigFilePath, err := cmd.Flags().GetString("ltagent-config")
	if err != nil {
		return err
	}
	ltConfig, err := loadtest.ReadConfig(ltConfigFilePath)
	if err != nil {
		return err
	}

	c, err := coordinator.New(cfg, *ltConfig, log)
	if err != nil {
		return fmt.Errorf("failed to create coordinator: %w", err)
	}

	done, err := c.Run()
	if err != nil {
		return fmt.Errorf("failed to run coordinator: %w", err)
	}

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-interruptChannel:
		if err := c.Stop(); err != nil {
			return fmt.Errorf("failed to stop coordinator: %w", err)
		}
	case <-done:
	}

	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:          "ltcoordinator",
		SilenceUsage: true,
		RunE:         RunCoordinatorCmdF,
		PreRunE:      initConfig,
	}
	rootCmd.PersistentFlags().StringP("config", "c", "", "path to the configuration file to use")
	rootCmd.PersistentFlags().StringP("ltagent-config", "l", "", "path to the load-test agent configuration file to use")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func initConfig(cmd *cobra.Command, args []string) error {
	configFilePath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	cfg, err := coordinator.ReadConfig(configFilePath)
	if err != nil {
		return err
	}

	logger.Init(&cfg.LogSettings)

	return nil
}
