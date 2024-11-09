/**
 * Helps with the travel calendar to provide better rendering of things
 */


document.addEventListener('DOMContentLoaded', function () {
    const travelDays = document.querySelectorAll('.travelDay');

    var grid = new gridjs.Grid({
        columns: [
            'Filer Name',
            {
                name: 'Departure Date',
                formatter: (cell) => {
                    return new Date(cell).toLocaleDateString();
                }
            },
            {
                name: 'Return Date',
                formatter: (cell) => {
                    return new Date(cell).toLocaleDateString();
                }
            }, 
            'Destination',
            {
                name: 'Congress Person',
            }
        ],
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
    console.log('grid', grid);

    grid.render(document.getElementById('wrapper'));


    travelDays.forEach(function (day) {
        const monthInt = parseInt(day.getAttribute('data-month'));
        const dayInt = parseInt(day.innerText);
        day.addEventListener('click', function () {
            console.log('clicked on day', dayInt, monthInt);
        });
    });
});