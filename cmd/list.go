package cmd

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "txl",
	Short: "Utility for TransferXL",
	Long:  `Command line interface for TransferXL.com`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	listCmd.Flags().StringVarP(&user, "user", "u", "", "user account")
	listCmd.Flags().StringVarP(&password, "password", "p", "", "password for account")

	listCmd.MarkFlagRequired("user")
	listCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List transfers",
	Long:  `List all transfers`,
	Run: func(cmd *cobra.Command, args []string) {

		transfers, err := listTransfers(user, password)
		if err != nil {
			fmt.Println("Bad request.")
		} else {
			if len(transfers) == 0 {
				fmt.Println("No transfers.")
			} else {
				// Sort transfers from newest to oldest
				sort.Slice(transfers, func(i, j int) bool {
					ti, _ := strconv.ParseInt(transfers[i].CreationDate, 10, 64)
					tj, _ := strconv.ParseInt(transfers[j].CreationDate, 10, 64)
					return ti > tj
				})

				list(transfers)
			}
		}
	},
}

func truncateString(str string, num int) (t string) {
	t = str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		t = str[0:num] + "..."
	}
	return
}

func list(transfers []Transfer) {
	const padding = 2
	w := tabwriter.NewWriter(os.Stdout, 0, 0, padding, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "Url\t Size\t Expiry\t Encrypted\t Name\t Message")
	for _, t := range transfers {
		expiry, _ := strconv.ParseInt(t.Expiry, 10, 64)
		tm := fmt.Sprint(time.Unix(expiry/1e9, 0).Format("2006-01-02 15:04"))
		fmt.Fprintf(w, "https://transferxl.com/%s\t %s\t %s\t %v\t %s\t %s\n", t.Shorturl, humanize.IBytes(uint64(t.Size)), tm, t.Encrypted, t.Filename, truncateString(t.Message, 24))
	}
	w.Flush()
}
