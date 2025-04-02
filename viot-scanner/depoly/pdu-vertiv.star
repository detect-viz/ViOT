load('time.star', 'time')

phase_map = {
    "Phase A": "L1",
    "Phase B": "L2",
    "Phase C": "L3",
}

def round2(x):
    return float(int(x * 100 + 0.5)) / 100.0

def apply(metric):
    metrics = []
    for k, v in metric.fields.items():
        
        # scale
        if k in ["voltage"]:
            v = v * 0.1
        elif k in ["current", "energy"]:
            v = v * 0.01

        # to float
        v = round2(float(v))
        
        # Phase bank rename
        tags = dict(metric.tags)
        bank = tags.get("bank", "")
        if bank in phase_map:
            bank = phase_map[bank]
        fields = {k: v}

        # new metric
        m = Metric(metric.name, tags, fields)
        
        # nanoseconds to seconds
        m.time = time.now().unix
        metrics.append(m)
    return metrics
