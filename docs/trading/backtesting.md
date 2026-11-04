# Backtesting guide

Synthesis of [Lean](lean.md), [Jesse](jesse.md), [Freqtrade](freqtrade.md) and [Hummingbot](hummingbot.md), checked against the current engine (`internal/trading/backtest.go`: next eligible event, configured costs and slippage, volume participation cap, conservative stop on OHLC ambiguity, 70/30 split).

## Fill model rules

State each rule in the report so results are reproducible.

1. Signals use closed candles only. Evaluate at candle close, fill at the **next** eligible open.
2. When a candle could hit both stop and target, assume the stop (conservative). Where the low or high came first is unknowable without a finer timeframe; when a finer series exists, use it for intra-candle order.
3. Gaps through a stop fill at the open, not at the stop price.
4. Costs apply on both sides. Slippage is explicit and never zero by default.
5. Participation is capped by bar volume; an order above the cap fills partially over later bars.
6. Resting passive orders fill only when price trades through them, not when it merely touches.

## Bias checklist

- **Look-ahead:** indicators and higher-timeframe merges must use only data completed before the decision time. Add a perturbation test: change future candles and assert earlier signals are unchanged.
- **Survivorship:** the universe must include securities that were later delisted; document when it does not.
- **Warm-up:** block orders until every indicator has enough history.
- **Data snooping:** the more parameters tried, the less any one result means. Record the number of trials.
- **Corporate actions:** adjust prices and volumes consistently and state the adjustment mode.

## Metrics

Report with stated definitions: total and annualized return, volatility, Sharpe and Sortino (state risk-free rate and frequency), maximum drawdown and its duration, Calmar, profit factor, expectancy, win rate, average win and loss, exposure, turnover, trade count, total costs, and a Buy and Hold baseline. Flag runs with too few trades (about fewer than 30) as statistically weak.

## Validation protocol

1. Split chronologically: train 60-70%, test 15-20%, validation 15-20%.
2. Tune on train, select on test, touch validation once.
3. Prefer walk-forward windows over a single split when data allows.
4. Add robustness statistics: block-bootstrap Monte Carlo on trades, cost stress (double costs and slippage), and deflated Sharpe or probability of backtest overfitting when many variants were tried.
5. Paper trading is the final out-of-sample test, and its results are compared with the backtest's expectation, not tuned against.

## Status in the current engine

Implemented and recorded in [ADR-0016](../adrs/0016-fill-rules-and-validation-protocol.md): resting
orders and partial fills over several bars, the train, test and validation partitions with a
once-per-version validation guard, walk-forward windows, Sharpe, Sortino, Calmar, SQN and exposure with
a stated risk-free rate, a block bootstrap, cost stress, the deflated Sharpe ratio, a look-ahead
perturbation test, a golden-file test of the metrics, and strategy parameters declared as data.

Still open: a fill model that walks the order book (needs book data), a parameter optimizer on top of
the parameter schema, and survivorship-aware universes.
