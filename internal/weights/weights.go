package weights

// From bandwidth to mwan3 weight
func FromBandwidth(mbps int) int {
    if mbps <= 0 { return 1 }
    return mbps
}

