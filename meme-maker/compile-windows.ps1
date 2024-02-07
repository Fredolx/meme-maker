$programPath = "./build/bin/meme-maker.exe"
go build -v -o $programPath

[string[]]$dlls = [System.Collections.ArrayList]@()
$lddOutput = & "C:\msys64\usr\bin\ldd.exe" $programPath

$lddOutput | ForEach-Object {
    if ($_ -match '.* => (.+) \(.*$') {
        $dllPath = $matches[1].Trim()
        $dllPath = $dllPath -replace '/c/', 'C:\' -replace '/', '\'
        if (-not ($dllPath -like "C:\Windows\*")) {
            if (-not $dllPath.StartsWith("C:")) {
                $dllPath = "C:\msys64" + $dllPath
            }
            $dlls += $dllPath
        }
    }
}

foreach ($dll in $dlls) {
    Write-Output $dll
    Copy-Item $dll -Destination ./bin
}

New-Item -Path "build" -Name "lib" -ItemType Directory
Copy-Item -Path "C:\msys64\mingw64\lib\ImageMagick-7.1.1" -Destination "build\lib" -Recurse
