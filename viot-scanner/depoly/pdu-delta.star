load('time.star', 'time')

def round2(x):
    return float(int(x * 100 + 0.5)) / 100.0

def apply(metric):
    metrics = []
    for k, v in metric.fields.items():
        if "." not in k:
            continue
        field, bank = k.split(".", 1)

        # scale
        if field in ["current", "voltage", "energy"]:
            v = v * 0.1

        # to float
        v = round2(float(v))
        
        # new metric
        tags = dict(metric.tags)
        tags["bank"] = bank
        fields = {field: v}
        m = Metric(metric.name, tags, fields)
        
        # nanoseconds to seconds
        m.time = time.now().unix
        metrics.append(m)
    return metrics
