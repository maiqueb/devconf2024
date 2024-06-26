package sql

func PersistIPQuery() string {
	return `insert into ips(pod_id, interface, ip) values($1, $2, $3)`
}

func DeleteIPQuery() string {
	return `
		with deleted_rows as (
			delete from ips where pod_id=$1 and interface=$2 returning *
		) select count(*) from deleted_rows;
`
}

func SelectAllQuery() string {
	return "select * from ips;"
}
