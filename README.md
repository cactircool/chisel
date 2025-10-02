# chisel

A highly optimized static parser generator.

Currently working except for a memory issue in how the tokens are returned. Currently deciding if the overhead of copying static token stack memory is worth it. Technically it is only a one byte copy, since static tokens hold no data, but its wrapped in a Token union, so no it would be at least a 9 byte copy.

Idk we'll see later.

Non-nested construct logic is proven to work however. I'll need to test speed compared to a bespoke parser later.
