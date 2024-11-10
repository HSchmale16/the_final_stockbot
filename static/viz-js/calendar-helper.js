/**
 * Helps with the travel calendar to provide better rendering of things
 */

// import * as gridjs from 'gridjs';

function renderMemberName(row) {
    const name = row.Member.CongressMemberInfo.name.official_full;
    const lastTerm = row.Member.CongressMemberInfo.terms[row.Member.CongressMemberInfo.terms.length - 1];

    if (lastTerm.type == "rep") {
        return `${name} (${lastTerm.party}-${lastTerm.state}-${lastTerm.district})`;
    }

    return `${name} (${lastTerm.party}-${lastTerm.state})`;
}

document.addEventListener('DOMContentLoaded', function () {
    const travelDays = document.querySelectorAll('.travelDay');

    var grid = new gridjs.Grid({
        pagination: true,
        search: true,
        columns: [
            'Filer Name',
            {
                name: 'Sponsor',
            },
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
            {
                name: 'Destination'
            },
            {
                name: 'Congress Person',
            }
        ],
        server: {
            url: `/json/travel/calendar/${yearInt}/${monthInt}`,
            then: data => data.map(row => [
                row.FilerName,
                row.TravelSponsor,
                row.DepartureDate,
                row.ReturnDate,
                row.Destination,
                gridjs.html(`<a href="/congress-member/${row.MemberId}">${renderMemberName(row)}</a>`)
            ])
        }
    })
    // console.log('grid', grid);
    // grid.plugin.add({
    //     id: 'magical-filter-bullshit',
    //     component: MagicalFilterBullshitPlugin,
    //     position: gridjs
    // })

    grid.render(document.getElementById('wrapper'));

    travelDays.forEach(function (day) {
        const monthInt = parseInt(day.getAttribute('data-month'));
        const dayInt = parseInt(day.innerText);
        day.addEventListener('click', function () {
            console.log('clicked on day', dayInt, monthInt);

            travelDays.forEach(function (day) {
                day.classList.remove('selected');
            });
        });
    });
});

function MagicalFilterBullshitPlugin() {
    return gridjs.h('div', {}, 'Hello World');
}