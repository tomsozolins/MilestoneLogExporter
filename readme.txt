
1) Install MilestonePSTools
PS# Install-Module -Name MilestonePSTools

Update if exists:
PS# Update-Module -Name MilestonePSTools

2) Test MilestonePSTools manually
# Connect-ManagementServer -Server localhost -AcceptEula; Get-Log -LogType Audit -BeginTime '2021-05-24 12:56:55' -EndTime '2021-05-24 12:57:55' | Foreach-Object { $_.'Message text' = $_.'Message text'.Replace("`r`n", " "); $_ } | Export-Csv -Path "C:\audit.csv" -NoTypeInformation -Encoding UTF8 ; Disconnect-ManagementServer

3) Install Log Worker service
CMD# CD <log-worker.exe directory path>
CMD# sc create log-worker binpath= "%CD%\log-worker.exe" start= auto DisplayName= "Log worker"
CMD# sc description log-worker "Log worker"

4) Change service account for Log Worker

5) Start Log Worker service

Export logs manually (only for days, do not change 00:00:00):
# cd <export.exe path>
# export "<management-server-address>" "Audit" "2021-05-01 00:00:00" "2021-05-31 00:00:00"
# export "<management-server-address>" "System" "2021-05-01 00:00:00" "2021-05-31 00:00:00"

//////////
Delete Log Worker service
CMD# sc delete log-worker
Check service
CMD# sc interrogate log-worker

Powershell TLS connection fix:
PS# [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12