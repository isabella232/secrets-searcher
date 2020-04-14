package dev

var (
    Enabled = false // Set this true by passing --dev

    EnableTestMode = false

    EnableCodePhase   = true
    EnableSearchPhase = true
    EnableReportPhase = true

    Repo         = ""
    Commit       = ""
    Path         = ""
    DiffLine     = -1
    LineContains = ""
    Processor    = ""
)

func init() {
    //EnableTestMode = true

    //EnableCodePhase = false
    //EnableSearchPhase = false
    //EnableReportPhase = false

    //Repo = "hermes"
    //Commit = "a69d35dc0612535167261e8ea8a8d61e9b7d2f76"
    //Path = "scripts/rotate-auth0-secrets.sh"
    //Processor = "generic-secret"
    //DiffLine = 16

    // Keep this here
    if EnableTestMode {
        Repo = ""
        Commit = ""
        Path = ""
        DiffLine = -1
        LineContains = ""
        Processor = ""
    }
}
