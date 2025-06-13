# CPU-Intensive Computation Example

This example demonstrates how to handle CPU-intensive computations in masc applications while keeping the UI responsive. It shows the difference between blocking and yielding computation patterns.

## The Problem

When a UI interaction triggers a CPU-intensive computation, the computation can block the UI thread, preventing visual updates until the computation completes. This creates a poor user experience where UI changes appear frozen.

## The Solution

**Yielding Control**: Periodically yield control back to the browser's event loop during long-running computations using:

- `masc.Yield()` - Frame-aware yielding function optimized for UI responsiveness

## What This Example Demonstrates

### Interactive Toggle
- **Checkbox control** to enable/disable yielding behavior
- **Real-time comparison** of blocking vs. yielding computations

### Radio Button Test
- **Three computation options** with different prime number targets (100k, 200k, 400k primes)
- **Immediate visual feedback** when yielding is enabled
- **Blocked UI updates** when yielding is disabled

### Visual Feedback
- **Status indicators** showing current computation state
- **Result display** with computation details
- **Clear instructions** for testing the difference

## Expected Behavior

### With Yielding Enabled ✅
1. Click a radio button
2. Radio button **immediately** shows as selected
3. "Selected Option" text **immediately** updates
4. Computing message appears
5. UI remains responsive during computation
6. Completion message appears when done

### With Yielding Disabled ❌
1. Click a radio button
2. Radio button selection **doesn't update** visually
3. "Selected Option" text **doesn't update**
4. UI appears frozen during computation
5. All updates appear **only after** computation completes

## Implementation Details

### Yielding Pattern
```go
for primeCount < target {
    // Do computation work
    if isPrime(num) {
        lastPrime = num
        primeCount++
    }
    num++
    
    // Yield control every 100000 iterations
    yieldCounter++
    if yieldCounter%100000 == 0 {
        masc.Yield() // Frame-aware yielding for optimal INP
    }
}
```

### Blocking Pattern
```go
for primeCount < target {
    // Do computation work
    if isPrime(num) {
        lastPrime = num
        primeCount++
    }
    num++
    // No yielding - blocks the UI thread
}
```

## Running the Example

```bash
masc serve ./example/computation
```

## Testing Instructions

1. **Start with yielding enabled** (default)
2. **Select different radio options** - notice immediate UI updates
3. **Disable yielding** using the checkbox
4. **Toggle back and forth** to see the dramatic difference

## When to Use Yielding

Use this pattern when:
- **Long-running computations** (>100ms)
- **UI responsiveness** is critical
- **Iterative algorithms** that can be broken into chunks
- **User interaction** triggers the computation

## Performance Considerations

- **Yielding frequency**: Every 100,000 iterations balances responsiveness with performance
- **Frame-aware timing**: `masc.Yield()` uses ~16ms delays optimized for browser rendering
- **Optimal INP**: Designed to achieve good Interaction to Next Paint performance
- **Computation speed**: Yielding adds minimal overhead while maintaining excellent UI responsiveness

## The masc.Yield() Function

The `masc.Yield()` function is specifically designed for CPU-intensive computations:

```go
func Yield() {
    runtime.Gosched()              // Yield to other goroutines
    time.Sleep(16 * time.Millisecond) // Wait ~1 frame for optimal INP
}
```

This ensures that:
- UI updates can complete within browser frame budgets
- Interaction to Next Paint (INP) metrics remain excellent
- Computation maintains good performance while staying responsive
