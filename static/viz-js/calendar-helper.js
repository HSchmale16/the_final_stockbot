/**
 * Helps with the travel calendar to provide better rendering of things
 */

document.addEventListener('DOMContentLoaded', function() {
    const travelDays = document.querySelectorAll('.travelDay');

    var grid = new gridjs.Grid({
        columns: ['Filer Name', 'Departure Date', 'Return Date', 'Destination', 'Congress Person'],
        server: {
            url: `/json/travel/calendar/${yearInt}/${monthInt}`,
            then: data => data.map(row => [
                row.FilerName,
                row.DepartureDate,
                row.ReturnDate,
                row.Destination,
                row.MemberName
            ])
        }
    })
    
    grid.render(document.getElementById('wrapper'));


    travelDays.forEach(function(day) {
        const monthInt = parseInt(day.getAttribute('data-month'));
        const dayInt = parseInt(day.innerText);
        day.addEventListener('click', function() {
            console.log('clicked on day', dayInt, monthInt);
        });
    });
});