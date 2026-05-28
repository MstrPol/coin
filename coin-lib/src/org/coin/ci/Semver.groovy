package org.coin.ci

class Semver {
    static List<Integer> parse(String v) {
        if (!v) return [0, 0, 0]
        def m = (v =~ /(\d+)\.(\d+)\.(\d+)/)
        if (!m.find()) return [0, 0, 0]
        return [m.group(1) as int, m.group(2) as int, m.group(3) as int]
    }

    static int compare(String a, String b) {
        def pa = parse(a)
        def pb = parse(b)
        for (int i = 0; i < 3; i++) {
            if (pa[i] != pb[i]) return pa[i] <=> pb[i]
        }
        return 0
    }
}

