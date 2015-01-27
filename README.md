# ConsumablePriceIndex

## Purpose
the ConsumablePriceIndex started out as sort of a _*Virtual* Consumer Price Index_ whose purpose quickly morphed into tracking prices and sales of specific consumables that my family uses regularly. the idea is to identify exactly when ordering these consumables online would be more cost effective than going to a local store. maintaining a historical record of consumable prices for research purposes remains as a secondary goal of the project.

## Setup
use the sample.config to put your specific data for both amazon and walmart api's. rename it to ~/.cpi/config.

the amazon Product API does not support IAM authentication, so we use the IAM role access for dynamodb to read encryption passwords from the db for the product api.
