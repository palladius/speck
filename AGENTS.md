I'm Riccardo usually I code with Agy/Gemini cLI so its my first non-gemini.md in my life! :)


# Auth

Note this claude session is wrapped by my $GIC/bin/claudio-XXXX scripts. 
1. There's one in python to test that model is working and non fa i capricci.
2. Then there's the 'claudio' wrapper which exportes the right ENV variables. Notably it points to a different JSON for ADC.
3. Then there's the authenticate-as-ricc which is needed to create/refresh the JSOn in a non-standard position.

## Specs

1. ensure you code things by following `docs/META-SPECS.md` and `docs/SPECS.md`
2. Do not change specs alone, always prompt user and ask for explicit confirmation to change those. if You do, do it in concise way (ie add 1-2 lines not a full paragraph or chapter!)
3. If you observe drift betwene code and specs, PLEASE let user know so you can take action.

## code

* Maintain a version somewhere (either in file or inside constants/main code file) and add changes to `CHANGELOG.md` consistently with version changes.
* Do not exit this git repo "cage" for coding. Seek approval if needed.
