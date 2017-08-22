# glint
custom lint tasks

### sub command helpers
It looks for helper functions and finds out that how many times they've got called.

Name convention is `_2helper()` should get called exactly two times and `_1helper()` should get called exactly one time. In this specific case, `_1helper()` can be `_helper()`.