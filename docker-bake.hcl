variable "V" {
    default = "latest"
}

group "default" {
    targets = ["gite-exporter"]
}

target "gite-exporter" {
    dockerfile = "Dockerfile"
    context = "."
    platforms = ["linux/amd64", "linux/arm64"]
    tags = ["docker.io/rockclimber81/gite-exporter:${V}"]
}