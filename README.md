# casetrack

A tracker of cases listed at [USDOJ: Investigations Regarding Violence at the Capitol](https://www.justice.gov/opa/investigations-regarding-violence-capitol).

## Additional

The case files reflect data that is available in the table at the above [justice.gov](https://www.justice.gov/opa/investigations-regarding-violence-capitol) webpage. This may be an incomplete listing of cases at any given moment. Some supplemental information (e.g. residency) has been manually added to the case files based on information available at the [Capitol Hill Cases](https://extremism.gwu.edu/Capitol-Hill-Cases) page provided by [GWU's Program on Extremism](https://extremism.gwu.edu/).

## Lists

Unique casenumbers:

```bash
jq -c '.[].casenumber' cases.json | sort | uniq -c
```

Unique names:

```bash
jq -c '.[].name' cases.json | sort | uniq -c
```

Names with missing casenumbers:
```bash
jq -c '.[] | select(.casenumber=="") | .name ' cases.json | sort
```
