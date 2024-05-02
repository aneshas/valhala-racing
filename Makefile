stripe-listen:
	@stripe listen --forward-to localhost:4000/checkout/payment-callback

boiler:
	@sqlboiler -c ./checkout/sqlboiler.toml psql
	@sqlboiler -c ./provisioner/sqlboiler.toml psql
