// Define function as global so that it can be run again
// after checkpoint and restart.
function test_fibonacci10k() {
  for (var run = 1; run <= 3; run++) {
    var start = Date.now();
    var fibonacci = function(n, output) {
      var a = 1, b = 1, sum;
      for (var i = 0; i < n; i++) {
        output.push(a);
        sum = a + b;
        a = b;
        b = sum;
      }
    }
    for(var i = 0; i < 10000; i++) {
      var result = [];
      fibonacci(78, result);
    }
    result;
    var ms = Date.now() - start;
    $.system.log('Run #' + run + ' fibonacci10k: ' + ms + 'ms');
  }
}

$.system.log('Benchmarking fibonacci10k...');
test_fibonacci10k();