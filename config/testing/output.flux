import "sql"


original_bucket = "raw"

measurement = "pdu"

sqlServer = "10.99.1.83:1433"

sqlAuth = "sa:1qaz2wsx"

sqlDB = "EX_SYSTEX"

tableName = "PowerCollectorValue_RawData_PDU"

raw =
    from(bucket: original_bucket)
        |> range(start: -10m)
        |> filter(fn: (r) => r["_measurement"] == measurement)
        |> filter(
            fn: (r) =>
                r["_field"] == "branch_current" or 
                r["_field"] == "branch_current" or
                r["_field"] == "branch_energy" or 
                r["_field"] == "branch_watt",
        )
        |> last()
        |> map(
            fn: (r) =>
                ({r with name:
                        string(v: r.factory) + 
                        string(v: r.phase) + 
                        string(v: r.datacenter) +
                        string(v: r.room) +
                        string(v: r.rack) + "P" + 
                        string(v: r.side),
                }),
        )
        |> map(fn: (r) => ({r with timestamp: int(v: r._stop) / 1000000000}))
        |> keep(
            columns: [
                "_field",
                "timestamp",
                "bank",
                "_value",
                "name",
            ],
        )
        |> rename(
            columns: {
                _field: "TYPE",
                _value: "VALUE",
                timestamp: "TIME_STAMP",
                branch: "BANK",
                name: "PDU_NAME",
            },
        )
        |> sql.to(
            driverName: "sqlserver",
            dataSourceName: "sqlserver://" + sqlAuth + "@" + sqlServer + "?database=" + sqlDB,
            table: tableName,
            batchSize: 1000,
        )
