### Install MilestonePSTools
```
Install-Module -Name MilestonePSTools
```

### Update MilestonePSTools
```
Update-Module -Name MilestonePSTools
```

### Test MilestonePSTools manually
```
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
Connect-ManagementServer -Server localhost -AcceptEula; Get-Log -LogType Audit -BeginTime '2021-05-24 12:56:55' -EndTime '2021-05-24 12:57:55' | Foreach-Object { $_.'Message text' = $_.'Message text'.Replace("`r`n", " "); $_ } | Export-Csv -Path "C:\audit.csv" -NoTypeInformation -Encoding UTF8 ; Disconnect-ManagementServer
```

### Install Log Worker service
```
cd <log-worker.exe directory path>
sc create log-worker binpath= "%CD%\log-worker.exe" start= auto DisplayName= "Log worker"
sc description log-worker "Log worker"
```
#### Change service account for Log Worker
#### Start Log Worker service

#### Export logs manually (do not change 00:00:00)
```
cd <export.exe path>
export.exe "<management-server-address>" "Audit" "2021-05-01 00:00:00" "2021-05-31 00:00:00"
export.exe "<management-server-address>" "System" "2021-05-01 00:00:00" "2021-05-31 00:00:00"
```

#### Delete Log Worker service
```
CMD# sc delete log-worker
```
#### Check service
```
CMD# sc interrogate log-worker
```
