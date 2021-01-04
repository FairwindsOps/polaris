$(function () {
  var data = [
    polarisSummary.Successes,
    polarisSummary.Warnings,
    polarisSummary.Dangers,
  ];
  var sum = data.reduce(function(total, cur) { return total + cur }, 0.0)
  if (sum === 0.0) {
    data = [1, 0, 0];
  }
  var clusterChart = new Chart("clusterScoreChart", {
    type: 'doughnut',
    data: {
      labels: ["Passing", "Warning", "Error"],
      datasets: [{
        data: data,
        backgroundColor: ['#8BD2DC', '#f26c21', '#a11f4c'],
      }]
    },
    options: {
      responsive: false,
      cutoutPercentage: 60,
      legend: {
        display: false,
      },
    }
  });
});
