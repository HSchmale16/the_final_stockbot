// /**
//  * Helps with the travel calendar to provide better rendering of things
//  */

// // import * as gridjs from 'gridjs';

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

    var table = new Tabulator("#wrapper", {
        ajaxURL: `/json/travel/calendar/${yearInt}/${monthInt}`,
        pagination: true,
        // progressiveLoad:"scroll",
        paginationSize:20,
        height: '500px',
        columns: [
            {
                title: "Filer Name",
                field: 'FilerName',
            },
            {
                title: 'Sponsor',
                field: 'TravelSponsor',
            },
            {
                title: 'Departure Date',
                field: 'DepartureDate',
                formatter: 'date',
            },
            {
                title: 'Return Date',
                field: 'ReturnDate',
                formatter: 'date',
            },
            {
                title: 'Destination',
                field: 'Destination'
            },
        ]
    })

    console.log('table', table);

    // var grid = new gridjs.Grid({
    //     pagination: true,
    //     search: true,
    //     columns: [
    //         'Filer Name',
    //         {
    //             name: 'Sponsor',
    //         },
    //         {
    //             name: 'Departure Date',
    //             formatter: (cell) => {
    //                 return new Date(cell).toLocaleDateString();
    //             }
    //         },
    //         {
    //             name: 'Return Date',
    //             formatter: (cell) => {
    //                 return new Date(cell).toLocaleDateString();
    //             }
    //         }, 
    //         {
    //             name: 'Destination'
    //         },
    //         {
    //             name: 'Congress Person',
    //         }
    //     ],
    //     server: {
    //         url: `/json/travel/calendar/${yearInt}/${monthInt}`,
    //         then: data => data.map(row => [
    //             row.FilerName,
    //             row.TravelSponsor,
    //             row.DepartureDate,
    //             row.ReturnDate,
    //             row.Destination,
    //             gridjs.html(`<a href="/congress-member/${row.MemberId}">${renderMemberName(row)}</a>`)
    //         ])
    //     }
    // })

});