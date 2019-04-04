$(function () {
  // Chart.js plugin for rendering text inside of donut chart
  // Credit: https://stackoverflow.com/a/43026361/8870697
  Chart.pluginService.register({
    beforeDraw: function (chart) {
      if (chart.config.options.elements.center) {
        var ctx = chart.chart.ctx;
        var centerConfig = chart.config.options.elements.center;
        var fontStyle = centerConfig.fontStyle || 'Arial';
        var txt = centerConfig.text;
        var color = centerConfig.color || '#000';
        var sidePadding = centerConfig.sidePadding || 20;
        var sidePaddingCalculated = (sidePadding/100) * (chart.innerRadius * 2)
        // Start with a base font of 30px
        ctx.font = "30px " + fontStyle;

        // Get the width of the string and also the width of the element minus 10 to give it 5px side padding
        var stringWidth = ctx.measureText(txt).width;
        var elementWidth = (chart.innerRadius * 2) - sidePaddingCalculated;

        // Find out how much the font can grow in width.
        var widthRatio = elementWidth / stringWidth;
        var newFontSize = Math.floor(30 * widthRatio);
        var elementHeight = (chart.innerRadius * 2);

        // Pick a new font size so it will not be larger than the height of label.
        var fontSizeToUse = Math.min(newFontSize, elementHeight);

        // Set font settings to draw it correctly.
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        var centerX = ((chart.chartArea.left + chart.chartArea.right) / 2);
        var centerY = ((chart.chartArea.top + chart.chartArea.bottom) / 2);
        ctx.font = "bold "+fontSizeToUse+"px " + fontStyle;
        ctx.fillStyle = color;

        // Draw text in center
        ctx.fillText(txt, centerX, centerY);
      }
    }
  });

  $('.namespace .resource-info .name').on('click', function(e) {
    console.log('clicked', arguments)
    console.log('parent', $(e.srcElement).parent('.resource-info'));
    $(e.srcElement).parents('.resource-info').toggleClass('expanded');
  });

  var expandMatch = window.location.search.match(/expand=(\w+)(\W|$)/);
  if (expandMatch && expandMatch[1] !== 'false' && expandMatch[1] !== '0') {
    $('.resource-info').addClass('expanded');
  }
});
