$(function () {
  var clusterChart = new Chart("clusterScoreChart", {
    type: 'doughnut',
    data: {
      labels: ["Passing", "Warning", "Error"],
      datasets: [{
        data: [
          polarisSummary.Successes,
          polarisSummary.Warnings,
          polarisSummary.Dangers,
        ],
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
