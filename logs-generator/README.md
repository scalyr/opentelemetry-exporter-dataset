# Logs Generator

Simple log generator tool that allows to generate specific number of lines with their length following normal distribution.


## Usage

1. Build docker image - `make docker-build`
2. Run with default parameters - `make docker-run`
3. Run with specified arguments - `make docker-run-with-args`

## Generator Parameters

Length of the line is sampled from the [normal distribution](https://www.wolframalpha.com/input?i2d=true&i=normal+distribution%5C%2840%29100%5C%2844%29+10%5C%2841%29) with the parameter mean `mu` and standard deviation `sigma`.
* `--lines` - number of lines to generatenon
* `--line-length-mu` - line length - `mu` parameter
* `--line-length-sigma` - line length - `sigma` parameter
* `--seed` - seed for random generator to get reproducibility

You can check the values for given parameters in [wolfram alpha](https://www.wolframalpha.com/input?i2d=true&i=normal+distribution%5C%2840%29100%5C%2844%29+10%5C%2841%29) or [here](https://homepage.divms.uiowa.edu/~mbognar/applets/normal.html).
