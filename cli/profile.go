package main

import (
	"fmt"
	"keentune/daemon/common/config"
	"strings"

	"github.com/spf13/cobra"
)

const (
	egInfo         = "\tkeentune profile info --name cpu_high_load.conf"
	egSet          = "\tkeentune profile set --group1 cpu_high_load.conf"
	egGenerate     = "\tkeentune profile generate --name tune_test.conf --output gen_param_test.json"
	egProfDelete   = "\tkeentune profile delete --name tune_test.conf"
	egProfList     = "\tkeentune profile list"
	egProfRollback = "\tkeentune profile rollback"
)

func createProfileCmds() *cobra.Command {
	var profCmd = &cobra.Command{
		Use:     "profile [command]",
		Short:   "Static tuning with expert profiles",
		Long:    "Static tuning with expert profiles",
		Example: fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", egProfDelete, egGenerate, egInfo, egProfList, egProfRollback, egSet),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if args[0] != "--help" && args[0] != "-h" && args[0] != "generate" && args[0] != "list" && args[0] != "set" && args[0] != "delete" && args[0] != "info" && args[0] != "rollback" {
					fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				}
			}

			if len(args) == 0 {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
			}

			return cmd.Help()
		},
	}

	var profileCommands []*cobra.Command
	profileCommands = append(profileCommands, decorateCmd(infoCmd()))
	profileCommands = append(profileCommands, decorateCmd(setCmd()))
	profileCommands = append(profileCommands, decorateCmd(deleteProfileCmd()))
	profileCommands = append(profileCommands, decorateCmd(listProfileCmd()))
	profileCommands = append(profileCommands, decorateCmd(rollbackCmd("profile")))
	profileCommands = append(profileCommands, decorateCmd(generateCmd()))

	profCmd.AddCommand(profileCommands...)
	return profCmd
}

func infoCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:     "info",
		Short:   "Show information of the specified profile",
		Long:    "Show information of the specified profile",
		Example: egInfo,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			name = strings.TrimSuffix(name, ".conf") + ".conf"
			RunInfoRemote(cmd.Context(), name)
			return
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "profile name, query by command \"keentune profile list\"")

	return cmd
}

func listProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all profiles",
		Long:    "List all profiles",
		Example: egProfList,
		Run: func(cmd *cobra.Command, args []string) {
			RunListRemote(cmd.Context(), "profile")
			return
		},
	}

	return cmd
}

// func setCmd() *cobra.Command {
// 	var setFlag SetFlag
// 	cmd := &cobra.Command{
// 		Use:     "set",
// 		Short:   "Apply a profile to the target machine",
// 		Long:    "Apply a profile to the target machine",
// 		Example: egSet,
// 		Run: func(cmd *cobra.Command, args []string) {
// 			if strings.Trim(setFlag.Name, " ") == "" {
// 				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
// 				cmd.Help()
// 				return
// 			}

// 			setFlag.Name = strings.TrimSuffix(setFlag.Name, ".conf") + ".conf"
// 			RunSetRemote(cmd.Context(), setFlag)
// 			return
// 		},
// 	}

// 	cmd.Flags().StringVar(&setFlag.Name, "name", "", "profile name, query by command \"keentune profile list\"")
// 	return cmd
// }

func setCmd() *cobra.Command {
	var setFlag SetFlag
	const GroupNum int = 20
	conf := new(config.KeentunedConf)
	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Apply a profile to the target machine",
		Long:    "Apply a profile to the target machine",
		Example: egSet,
		Run: func(cmd *cobra.Command, args []string) {
			/*
				if strings.Trim(setFlag.Name, " ") == "" {
					fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
					cmd.Help()
					return
				}
				setFlag.Name = strings.TrimSuffix(setFlag.Name, ".conf") + ".conf"
			*/
			fmt.Println("args", args)
			//判断若args有值且以.conf结尾，则认为是默认所有group下发统一配置
			if len(args) > 0 && strings.HasSuffix(args[0], ".conf") {
				for i, _ := range setFlag.ConfFile {
					setFlag.Group[i] = true
					setFlag.ConfFile[i] = args[0]
				}
			} else {
				//若groupX已配置且以.conf结尾，则认为该配置有效
				for i, v := range setFlag.ConfFile {
					if len(v) != 0 && strings.HasSuffix(v, ".conf") {
						setFlag.Group[i] = true
					} else {
						setFlag.Group[i] = false
					}
				}
			}
			// for _, v := range setFlag.ConfFile {
			// 	fmt.Println(v)
			// }
			// for _, v := range setFlag.Group {
			// 	fmt.Println(v)
			// }
			RunSetRemote(cmd.Context(), setFlag)
			return
		},
	}

	var group string = ""
	if err := conf.Save(); err != nil {
		setFlag.Group = make([]bool, GroupNum)
		setFlag.ConfFile = make([]string, GroupNum)
		for index := 0; index < GroupNum; index++ {
			group = fmt.Sprintf("group%d", index)
			cmd.Flags().StringVar(&setFlag.ConfFile[index], group, "", "profile name, query by command \"keentune profile list\"")
		}
	} else {
		/*
			setFlag.Group = make([]string, len(conf.TargetIP))
			setFlag.ConfFile = make([]string, len(conf.TargetIP))
			for index, _ := range conf.TargetIP {
				group = fmt.Sprintf("group%d",index+1)
				cmd.Flags().StringVar(&setFlag.Group[index], group, "", "profile name, query by command \"keentune profile list\"")
			}
		*/

		setFlag.Group = make([]bool, GroupNum+2)
		setFlag.ConfFile = make([]string, GroupNum+2)
		for index := 0; index < GroupNum; index++ {
			group = fmt.Sprintf("group%d", index+1)
			cmd.Flags().StringVar(&setFlag.ConfFile[index], group, "", "profile name, query by command \"keentune profile list\"")
		}

	}

	return cmd
}

func deleteProfileCmd() *cobra.Command {
	var flag DeleteFlag
	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a profile",
		Long:    "Delete a profile",
		Example: egProfDelete,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(flag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			flag.Cmd = "profile"
			flag.Name = strings.TrimSuffix(flag.Name, ".conf") + ".conf"
			RunDeleteRemote(cmd.Context(), flag)
			return
		},
	}

	cmd.Flags().StringVar(&flag.Name, "name", "", "profile name, query by command \"keentune profile list\"")

	return cmd
}

func generateCmd() *cobra.Command {
	var genFlag GenFlag
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Generate a parameter configuration file from profile",
		Long:    "Generate a parameter configuration file from profile",
		Example: egGenerate,
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Trim(genFlag.Name, " ") == "" {
				fmt.Printf("%v Incomplete or Unmatched command.\n\n", ColorString("red", "[ERROR]"))
				cmd.Help()
				return
			}

			genFlag.Name = strings.TrimSuffix(genFlag.Name, ".conf") + ".conf"
			if strings.Trim(genFlag.Output, " ") == "" {
				genFlag.Output = strings.TrimSuffix(genFlag.Name, ".conf") + ".json"
			} else {
				genFlag.Output = strings.TrimSuffix(genFlag.Output, ".json") + ".json"
			}

			RunGenerateRemote(cmd.Context(), genFlag)
			return
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&genFlag.Name, "name", "n", "", "profile name, query by command \"keentune profile list\"")
	flags.StringVarP(&genFlag.Output, "output", "o", "", "output parameter configuration file name, default with suffix \".json\"")

	return cmd
}
