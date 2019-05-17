# comed_exporter - turn comed hourly pricing data into prometheus metrics

ComEd offers hourly pricing to some customers.  They provide an api into this data at https://hourlypricing.comed.com/hp-api/.

This exporter reads that data on a schedule.
